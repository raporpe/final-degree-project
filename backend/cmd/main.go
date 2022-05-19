package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/klauspost/oui"
	"github.com/raporpe/dbscan"
)

// The mac vendor IEEE database
var macDB oui.StaticDB

// Database initialization
var gormDB *gorm.DB

// Window size
var windowSizeSeconds = 60

func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "tfg-server.raporpe.dev", 5432, "postgres", "raulportugues", "tfg")

	var err error
	gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	gormDB.AutoMigrate(&DetectedMacDB{})
	gormDB.AutoMigrate(&PersonalMacsDB{})

	r := mux.NewRouter()
	r.HandleFunc("/v1/detected-macs", DetectedMacsPostHandler).Methods("POST")
	r.HandleFunc("/v1/detected-macs", DetectedMacsGetHandler).Methods("GET")
	r.HandleFunc("/v1/personal-macs", PersonalMacsHandler)
	r.HandleFunc("/v1/digested-macs", DigestedMacsHandler).Methods("GET")
	r.HandleFunc("/v1/clustered-macs", GetClusteredMacsHandler).Methods("GET")
	r.HandleFunc("/v1/last-room", GetLastRoomHandler).Methods("GET")
	r.HandleFunc("/v1/historic", GetHistoricHandler).Methods("GET")
	r.HandleFunc("/v1/config", ConfigGetHandler)

	serverPort := os.Getenv("API_PORT")

	// Default server port
	if serverPort == "" {
		serverPort = "2000"
	}

	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + serverPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server on port " + serverPort)

	go PeriodicRoomJob()

	CheckError(server.ListenAndServe())

}

func ConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving configuration")

	configResponse := ConfigResponse{
		SecondsPerWindow: 60,
	}

	byteJson, err := json.Marshal(configResponse)
	CheckError(err)
	w.Write(byteJson)

}

func DigestedMacsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving mac digest")

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// Get the start time and end time
	startTime, endTime := GetStartEndTime()

	// If the start time and the end time are set, then override defaults
	startTimeSet := r.URL.Query().Get("start_time") != ""
	endTimeSet := r.URL.Query().Get("end_time") != ""

	if endTimeSet || startTimeSet {
		var err error
		// Read the query param start_time
		startTime, err = time.Parse(time.RFC3339, r.URL.Query().Get("start_time"))
		if err != nil {
			log.Println("Invalid start_time!: " + err.Error())
			w.WriteHeader(500)
			w.Write([]byte("Invalid start_time"))
			return
		}

		// Read the query param end_time
		endTime, err = time.Parse(time.RFC3339, r.URL.Query().Get("end_time"))
		if err != nil {
			log.Println("Invalid end_time!")
			w.WriteHeader(500)
			w.Write([]byte("Invalid end_time"))
			return
		}

	}

	// Read the query param device_id
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		log.Println("Invalid device_id!")
		w.WriteHeader(500)
		w.Write([]byte("device_id required"))
		return
	}

	digestedMacs := GetDigestedMacs(deviceID, startTime, endTime)

	jsonResponse, err := json.Marshal(&digestedMacs)
	if err != nil {
		w.WriteHeader(500)
		log.Println("There was an error trying to marshall the final struct!")
		return
	}

	w.Write([]byte(jsonResponse))

}

func GetLastRoomHandler(w http.ResponseWriter, r *http.Request) {
	rooms := GetRooms(GetLastTime())
	jsonResponse, err := json.Marshal(&rooms)
	if err != nil {
		w.WriteHeader(500)
		log.Println("There was an error trying to marshall the final struct!")
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Write([]byte(jsonResponse))
}

func GetHistoricHandler(w http.ResponseWriter, r *http.Request) {

	fromTimeSet := r.URL.Query().Get("from_time") != ""
	toTimeSet := r.URL.Query().Get("to_time") != ""

	var fromTime time.Time
	var toTime time.Time
	var err error

	if fromTimeSet {
		// Read the query param from_time
		fromTime, err = time.Parse(time.RFC3339, r.URL.Query().Get("from_time"))
		if err != nil {
			log.Println("Invalid from_time!: " + err.Error())
			w.WriteHeader(500)
			w.Write([]byte("Invalid from_time"))
			return
		}
	} else {
		w.WriteHeader(500)
		w.Write([]byte("If you set to time, from time is required"))
		return
	}

	if toTimeSet {
		// Read the query param to_time
		toTime, err = time.Parse(time.RFC3339, r.URL.Query().Get("to_time"))
		if err != nil {
			log.Println("Invalid to_time!")
			w.WriteHeader(500)
			w.Write([]byte("Invalid to_time"))
			return
		}
	} else {
		toTime = GetLastTime()
	}

	room, err := GetHistoricRoomInDB(fromTime, toTime)
	if err != nil {
		w.WriteHeader(500)
		log.Println("There was an error getting last room in database")
		return
	}

	jsonResponse, err := json.Marshal(&room)
	if err != nil {
		w.WriteHeader(500)
		log.Println("There was an error trying to marshall the final struct!")
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Write([]byte(jsonResponse))
}

func PeriodicRoomJob() {

	// Get last time and compare with database
	lastTime := GetLastTime()
	db, _ := GetLastRoomInDB() // Get last one

	// If the current time has alredy been processed
	if lastTime == db.EndTime {
		sleepTime := 60 - time.Now().Second() + 30 // Extra 30 seconds to have half a minute extra
		fmt.Printf("sleepTime: %v\n", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	} else {
		log.Println("Storing room data...")
		StoreRoomInDB(GetRooms(GetLastTime()))
	}

	// Infinite loop
	for {
		log.Println("Storing room data...")
		StoreRoomInDB(GetRooms(GetLastTime()))

		sleepTime := 60 - time.Now().Second() + 30 // Extra 30 seconds to have half a minute extra
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

}

func StoreRoomInDB(r ReturnRooms) error {
	// Generate the uuid for the room date
	uuid, err := uuid.NewUUID()
	if err != nil {
		return errors.New("Cannot generate UUID")
	}

	// Get locale
	l, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return errors.New("Cannot load time locale")
	}

	// Generate encoded data to store in db
	data, err := json.Marshal(r)
	if err != nil {
		return errors.New("Cannot encode data")
	}

	gormDB.Create(&RoomHistoricDB{
		ID:   uuid,
		Date: time.Now().In(l),
		Data: string(data),
	})

	return nil
}

func GetLastRoomInDB() (ReturnRooms, error) {
	var roomInDB RoomHistoricDB
	gormDB.Order("date DESC").Find(&roomInDB)

	// Decode the data that is stored in the database
	var room ReturnRooms

	err := json.Unmarshal([]byte(roomInDB.Data), &room)
	if err != nil {
		return ReturnRooms{}, errors.New("Error deserializing data stored in DB. Possible data corruption!")
	}

	return room, nil
}

func GetHistoricRoomInDB(from time.Time, to time.Time) (ReturnHistoricRooms, error) {

	var roomsInDB []RoomHistoricDB

	gormDB.Where("date >= ? and date <= ?", from, to).Order("date DESC").Find(&roomsInDB)

	var rooms []ReturnRooms
	for _, v := range roomsInDB {
		var room ReturnRooms

		err := json.Unmarshal([]byte(v.Data), &room)
		if err != nil {
			return ReturnHistoricRooms{}, errors.New("Error deserializing data stored in DB. Possible data corruption!")
		}

		rooms = append(rooms, room)

	}

	// Convert from type ReturnRooms to ReturnHistoricRooms
	ret := ReturnHistoricRooms{
		FirstTime: from,
		LastTime:  to,
		Rooms:     make(map[string]map[time.Time]int),
	}
	for _, v := range rooms {
		for r, i := range v.Rooms {
			if _, exists := ret.Rooms[r]; !exists {
				ret.Rooms[r] = make(map[time.Time]int)
			}
			ret.Rooms[r][v.EndTime] = i
		}
	}

	return ret, nil

}

func GetRooms(lastTime time.Time) ReturnRooms {
	// Get all the current rooms
	var allRooms []RoomsDB
	gormDB.Find(&allRooms)

	// Results: room (string) -> ocupation (int)
	rooms := make(map[string]int)

	// Bool for storing the rooms that are inconsistent
	inconsistentRooms := []string{}

	for _, v := range allRooms {

		// Get all the clusters from all the devices in the room up until time t
		clusteredMacs, err := GetClusteredMacs(v.RoomID, lastTime)
		if err != nil {
			fmt.Printf("There was an error getting room %v: %v", v.RoomID, err.Error())
		}

		// Check if the room has any inconsistent data
		if clusteredMacs.InconsistentData {
			inconsistentRooms = append(inconsistentRooms, v.RoomID)
		}

		// Iterate every device in the room and merge the clusters of each device
		var clusters [][]string

		// Generate general mappign
		// For each device in the room...
		for _, clustersOnDevice := range clusteredMacs.Results {
			clusters = ClusterMerge(clusters, clustersOnDevice, 0.33)
		}

		// Store how many clusters are in the room
		// clusters = people
		rooms[v.RoomID] = len(clusters)

	}

	return ReturnRooms{
		InconsistentRooms: inconsistentRooms,
		EndTime:           lastTime,
		StartTime:         lastTime.Add(-15 * time.Minute),
		ContextSize:       15,
		Rooms:             rooms,
	}
}

func GetClusteredMacsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving mac clustering")

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// Read the query param device_id
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		log.Println("Invalid room_id!")
		w.WriteHeader(500)
		w.Write([]byte("Ivalid room_id"))
		return
	}

	// Override end_time (optional)
	endTime := r.URL.Query().Get("end_time")
	var parsedEndTime time.Time
	if endTime != "" {
		var err error
		parsedEndTime, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			log.Println("Invalid end_time!")
			w.WriteHeader(500)
			w.Write([]byte("The specified end_time format is not valid"))
			return
		}
	}

	clusteredMacs, err := GetClusteredMacs(roomID, parsedEndTime)
	if err != nil {
		log.Println("Error calculating cluster: " + err.Error())
		w.WriteHeader(500)
		w.Write([]byte("An error ocurred"))
		return
	}

	jsonResponse, err := json.Marshal(&clusteredMacs)
	if err != nil {
		w.WriteHeader(500)
		log.Println("There was an error trying to marshall the final digested macs struct!")
		return
	}

	w.Write([]byte(jsonResponse))

}

func GetClusteredMacs(roomID string, endTime time.Time) (ReturnClusteredMacs, error) {

	// Get the devices that are in the room
	var CaptureDevicesInRoom []RoomsDB
	gormDB.Where("room_id = ?", roomID).Find(&CaptureDevicesInRoom)

	// Get the start time and end time
	var startTime time.Time
	// If the end time is not specified
	if endTime.IsZero() {
		startTime, endTime = GetStartEndTime()
	} else {
		// When the endTime is set, we calculate the start time
		startTime = endTime.Add(-15 * time.Minute)
	}

	// Create the return variable, a map from devices to clusters
	results := make(map[string][][]string)

	// To note if there was any inconsistency and return it at the end
	inconsistentData := false

	// Iterate over every device in the room
	for _, device := range CaptureDevicesInRoom {
		deviceID := device.DeviceID
		digestedMacs := GetDigestedMacs(deviceID, startTime, endTime)

		// Check for inconsistent data
		inconsistentData = inconsistentData || digestedMacs.InconsistentData

		// Exclude mac addresses that are not active
		for k, v := range digestedMacs.Digest {
			if !IsDeviceActive(v.PresenceRecord) {
				delete(digestedMacs.Digest, k)
			}
		}

		// From map to array
		var analyse []MacDigest
		var noAnalyse []MacDigest
		for _, v := range digestedMacs.Digest {
			// If the packet does not have a manufacturer, analyse it
			if v.Manufacturer == nil {
				analyse = append(analyse, v)
			} else {
				// If the packet does have a manufacturer,
				// no clustering is needed
				noAnalyse = append(noAnalyse, v)
			}
		}

		// Get the clusters from the analysis
		clusters := SimilarDetector(analyse)

		fmt.Printf("Analyzed clusters: %v\n", len(clusters))
		//fmt.Printf("clusters: %v\n", clusters)

		// Append to the clusters the macs that were not analyzed
		for _, v := range noAnalyse {
			// The not analyzed macs are true, so they are a cluster by themselves
			clusters = append(clusters, []string{v.Mac})
		}

		fmt.Printf("Analyzed + non analyzed clusters: %v\n", len(clusters))
		//fmt.Printf("Analyzed + non analyzed clusters result: %v\n", clusters)

		results[deviceID] = clusters

	}

	toReturn := ReturnClusteredMacs{
		StartTime:        startTime,
		EndTime:          endTime,
		InconsistentData: inconsistentData,
		Results:          results,
	}

	return toReturn, nil

}

func GetDigestedMacs(deviceID string, startTime time.Time, endTime time.Time) ReturnDigestedMacs {

	// Get data from the database
	var windowsInDB []DetectedMacDB
	gormDB.Where("device_id = ? and start_time >= ? and start_time < ? ", deviceID, startTime, endTime).Find(&windowsInDB)

	//.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.
	// PHASE 1 -> Data consistency check
	//.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.

	// Generate the start time of all the windows that should be in DB
	var expectedStartTimes []time.Time
	for expectedTime := startTime; expectedTime.Before(endTime); expectedTime = timePlusWindow(expectedTime, 1, windowSizeSeconds) {
		expectedStartTimes = append(expectedStartTimes, expectedTime)
	}

	// How many windows there should be
	expectedWindowsBetween := len(expectedStartTimes)

	// The number of windows that are in the database for the device and time range
	realWindowsBetween := len(windowsInDB)

	// Check that the number of windows in DB are the expected ones
	if realWindowsBetween != expectedWindowsBetween {
		log.Printf("Window mismatch!!. There are %v but %v were expected", realWindowsBetween, expectedWindowsBetween)
		log.Printf("realWindowsBetween: %v\n", len(windowsInDB))
		log.Printf("windowsBetween: %v\n", expectedWindowsBetween)
	}

	// Check that no window is repeated or missing
	inconsistentData := false
	inconsistentTimes := make([]time.Time, 0)

	for _, checkingTime := range expectedStartTimes {
		matchInDB := false
		multipleMatchInDB := false

		// For every expected time there exists a window in DB with that same start time
		for _, checkingWindow := range windowsInDB {
			if checkingWindow.StartTime.Equal(checkingTime) {

				// The first time that matches
				if !matchInDB {
					matchInDB = true
					log.Printf("Correct match! %v\n", checkingTime)
				} else {
					// If it had already matched
					multipleMatchInDB = true
					log.Printf("Double match! %v\n", checkingTime)
				}
			}
		}

		if !matchInDB || multipleMatchInDB {
			inconsistentData = true
			inconsistentTimes = append(inconsistentTimes, checkingTime)
			log.Printf("There is an inconsistency for date %v, match:%v, multipleMatch:%v\n", checkingTime, matchInDB, multipleMatchInDB)
		}

	}

	if inconsistentData {
		log.Println("There is an inconsistency in the data!")
	}

	//.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.
	// PHASE 2 -> Generate information that will be returned
	//.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.oOo.

	// Where the digested macs will be stored
	digestedMacs := make(map[string]MacDigest)

	// Generate the data for every expected time
	for _, expectedTime := range expectedStartTimes {

		// If the current expectedTime was marked inconsistent, skip it
		skip := false
		for _, i := range inconsistentTimes {
			if i.Equal(expectedTime) {
				skip = true
				//log.Printf("SKIPPING windo w.StartTime: %v\n", expectedTime)
			}
		}

		if skip {
			//log.Printf("SKIPPING window.StartTime: %v\n", expectedTime)
			continue
		}

		// Get the equivalent window in DB that has the expected time
		// We guaranteed in the data consistency check that exactly one will exist
		var window DetectedMacDB
		for _, w := range windowsInDB {
			if w.StartTime.Equal(expectedTime) {
				window = w
			}
		}

		currentWindowNumber := int(window.StartTime.Sub(startTime).Seconds()) / windowSizeSeconds
		//log.Printf("currentWindowNumber: %v\n", currentWindowNumber)

		// Generate the mac metadata struct back from db
		var macMetadata map[string]MacMetadata
		err := json.Unmarshal([]byte(window.DetectedMacs), &macMetadata)
		if err != nil {
			log.Printf("Possible data corruption in db for window id %v ", window.ID)
			// Skip this window since information is not reliable in db
			continue
		}

		// Iterate over all the detected mac addresses
		for mac, data := range macMetadata {
			// If the mac exists in digestedMacs
			if m, exists := digestedMacs[mac]; exists {
				c := howManyTrue(m.PresenceRecord)
				m.AvgSignalStrength = ((digestedMacs[mac].AvgSignalStrength * float64(c)) + data.AverageSignalStrength) / float64(c+1)
				m.PresenceRecord[currentWindowNumber] = true
				m.TypeCount[0] += data.TypeCount[0]
				m.TypeCount[1] += data.TypeCount[1]
				m.TypeCount[2] += data.TypeCount[2]
				m.SSIDProbes = DeduplicateSlice(append(m.SSIDProbes, data.SSIDProbes...))
				m.HTCapabilities = data.HTCapabilities
				m.HTExtendedCapabilities = data.HTExtendedCapabilities
				m.SupportedRates = DeduplicateSlice(append(m.SupportedRates, data.SupportedRates...))
				m.Tags = DeduplicateSlice(append(m.Tags, data.Tags...))

				// Assign the modified struct to the digested macs
				digestedMacs[mac] = m
			} else {
				// If the mac does no exist, create struct
				m := MacDigest{
					Mac:                    mac,
					AvgSignalStrength:      data.AverageSignalStrength,
					PresenceRecord:         make([]bool, expectedWindowsBetween),
					TypeCount:              data.TypeCount,
					Manufacturer:           GetMacVendor(mac),
					OuiID:                  GetMacPrefix(mac),
					SSIDProbes:             DeduplicateSlice(data.SSIDProbes),
					HTCapabilities:         data.HTCapabilities,
					HTExtendedCapabilities: data.HTExtendedCapabilities,
					SupportedRates:         DeduplicateSlice(data.SupportedRates),
					Tags:                   DeduplicateSlice(data.Tags),
				}
				// Set the presence record to true
				m.PresenceRecord[currentWindowNumber] = true

				// Assign new struct
				digestedMacs[mac] = m
			}

		}

	}

	digest := ReturnDigestedMacs{
		NumberOfWindows:   expectedWindowsBetween,
		WindowsStartTimes: expectedStartTimes,
		StartTime:         startTime,
		EndTime:           timePlusWindow(startTime, expectedWindowsBetween, windowSizeSeconds),
		Digest:            digestedMacs,
		InconsistentData:  inconsistentData,
		InconsistentTimes: inconsistentTimes,
	}

	// Enrich data with stored metadata before returning
	for mac, v := range digest.Digest {
		meta, err := GetPersonalMacMetadata(mac)
		if err != nil {
			// Skip current mac
			log.Printf("Error in mac %v when getting related metadata: %v", mac, err.Error())
		}

		// Merge ssid probes
		if len(v.SSIDProbes) > 0 {
			meta.AddSSID(v.SSIDProbes...)
			v.SSIDProbes = meta.SSIDProbes
		}

		// Set the ht capabilities
		if v.HTCapabilities != nil {
			meta.HTCapabilities = v.HTCapabilities
		} else {
			v.HTCapabilities = meta.HTCapabilities
		}

		// Set the extended ht capabilities
		if v.HTExtendedCapabilities != nil {
			meta.HTExtendedCapabilities = v.HTExtendedCapabilities
		} else {
			v.HTExtendedCapabilities = meta.HTExtendedCapabilities
		}

		// Set the supported rates
		if len(v.SupportedRates) > 0 {
			meta.SupportedRates = v.SupportedRates
		} else {
			v.SupportedRates = meta.SupportedRates
		}

		// Set the tags
		if len(v.Tags) > 0 {
			meta.Tags = v.Tags
		} else {
			v.Tags = meta.Tags
		}

		// Save the possible new values added to the metadata store
		meta.UpdateInDB(mac)

	}

	// Return the digested macs
	return digest

}

func PersonalMacsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	CheckError(err)

	var query PersonalMacsUpload

	err = json.Unmarshal(body, &query)
	CheckError(err)

	log.Println("Syncing config with " + query.DeviceID)
	var ret []PersonalMacsDB
	if r.Method == "POST" {
		// Just insert the mac addresses
		for _, mac := range query.PersonalMacs {
			gormDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&PersonalMacsDB{
				Mac: mac,
			})
		}
	}

	gormDB.Find(&ret)
	var response []string
	for _, elem := range ret {
		response = append(response, elem.Mac)
	}
	jsonResponse, err := json.Marshal(response)
	CheckError(err)
	w.Write(jsonResponse)

}

func DetectedMacsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	startTime, err := time.Parse(time.RFC3339, r.URL.Query().Get("start_time"))
	if err != nil {
		log.Println("Invalid start_time!: " + err.Error())
		w.Write([]byte("Invalid start_time"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, r.URL.Query().Get("end_time"))
	if err != nil {
		log.Println("Invalid end_time!")
		w.Write([]byte("Invalid end_time"))
		return
	}

	// Get data from the database
	var ret []DetectedMacDB
	gormDB.Where("start_time >= ? and start_time < ?", startTime, endTime).Find(&ret)

	response := make(map[time.Time]map[string]ReturnDetectedMacs)
	for _, elem := range ret {
		var d map[string]MacMetadata
		err := json.Unmarshal([]byte(elem.DetectedMacs), &d)
		CheckError(err)

		if response[elem.EndTime] == nil {
			response[elem.EndTime] = make(map[string]ReturnDetectedMacs)
		}

		response[elem.EndTime][elem.DeviceID] = ReturnDetectedMacs{
			DetectedMacs:     d,
			SecondsPerWindow: elem.SecondsPerWindow,
			EndTime:          elem.EndTime,
		}
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Error generating json with data from the Detected Macs table!")
		w.Write([]byte("There was an error serializing request"))
	}
	w.Write(jsonResponse)

}

func DetectedMacsPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var state UploadDetectedMacs

	err = json.Unmarshal(body, &state)
	if err != nil {
		log.Println("Error when unmarshalling received body: " + err.Error())
		w.WriteHeader(500)
		return
	}

	log.Println("Storing detected macs from " + state.DeviceID)

	// Generate UUID
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Print("Error generating UUID for inserting the detected macs! Skip this insertion: " + err.Error())
		w.WriteHeader(500)
		return
	}

	detectedMacs, err := json.Marshal(state.DetectedMacs)
	if err != nil {
		log.Print("Error remarshalling the detected macs! Can this even happen?")
		return
	}

	// Store the list of detected macs in the DB
	gormDB.Create(&DetectedMacDB{
		ID:               uuid,
		DeviceID:         state.DeviceID,
		SecondsPerWindow: state.SecondsPerWindow,
		StartTime:        time.Unix(int64(state.StartTime), 0),
		EndTime:          time.Unix(int64(state.EndTime), 0),
		DetectedMacs:     string(detectedMacs),
	})

}

func GetPersonalMacMetadata(mac string) (PersonalMacMetadata, error) {
	var ret PersonalMacsDB
	gormDB.Where("mac = ?", mac).Find(&ret)

	var metadata PersonalMacMetadata
	err := json.Unmarshal([]byte(ret.Metadata), &metadata)
	if err != nil {
		return PersonalMacMetadata{}, errors.New("An error ocurred retrieving personal mac metadata: " + err.Error())
	}

	return metadata, nil
}

func GetMacVendor(mac string) *string {
	if macDB == nil {
		var err error
		macDB, err = oui.OpenStaticFile("oui.txt")
		CheckError(err)
	}

	result, err := macDB.Query(mac)
	if err != nil {
		return nil
	} else {
		return &result.Manufacturer
	}
}

func GetMacPrefix(mac string) *string {
	if macDB == nil {
		var err error
		macDB, err = oui.OpenStaticFile("oui.txt")
		CheckError(err)
	}

	result, err := macDB.Query(mac)
	if err != nil {
		return nil
	} else {
		res := result.Prefix.String()
		return &res
	}
}

func CheckError(err error) {
	if err != nil {
		log.Print(err)
	}
}

type UploadDetectedMacs struct {
	DeviceID         string                 `json:"device_id"`
	DetectedMacs     map[string]MacMetadata `json:"detected_macs"`
	SecondsPerWindow int                    `json:"seconds_per_window"`
	StartTime        int                    `json:"start_time"`
	EndTime          int                    `json:"end_time"`
}

type DetectedMacDB struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;" json:"-"`
	DeviceID         string    `json:"device_id"`
	DetectedMacs     string    `json:"detected_macs"`
	SecondsPerWindow int       `json:"seconds_per_window"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
}

func (DetectedMacDB) TableName() string {
	return "detected_macs"
}

type RoomHistoricDB struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"-"`
	Date time.Time
	Data string
}

func (RoomHistoricDB) TableName() string {
	return "room_historic"
}

type PersonalMacMetadata struct {
	SSIDProbes             []string
	HTCapabilities         *string
	HTExtendedCapabilities *string
	SupportedRates         []float64
	Tags                   []int
}

func (p PersonalMacMetadata) UpdateInDB(mac string) {
	// Convert metadata to string
	meta, err := json.Marshal(p)
	if err != nil {
		log.Println("There was an error storing mac metadata: " + err.Error())
		return // Ommit error
	}

	var db PersonalMacsDB
	gormDB.Where("mac = ?", mac).Find(&db).Update("metadata", string(meta))

}

func (p *PersonalMacMetadata) AddSSID(ssid ...string) {
	p.SSIDProbes = DeduplicateSlice(append(p.SSIDProbes, ssid...))
}

type ReturnDigestedMacs struct {
	NumberOfWindows   int                  `json:"number_of_windows"`
	WindowsStartTimes []time.Time          `json:"windows_start_times"`
	StartTime         time.Time            `json:"start_time"`
	EndTime           time.Time            `json:"end_time"`
	InconsistentData  bool                 `json:"inconsistent_data"`
	InconsistentTimes []time.Time          `json:"inconsistent_times"`
	Digest            map[string]MacDigest `json:"digest"`
}

type ReturnDetectedMacs struct {
	DeviceID         string                 `json:"device_id"`
	DetectedMacs     map[string]MacMetadata `json:"detected_macs"`
	SecondsPerWindow int                    `json:"seconds_per_window"`
	EndTime          time.Time              `json:"end_time"`
}

type ReturnClusteredMacs struct {
	StartTime        time.Time             `json:"start_time"`
	EndTime          time.Time             `json:"end_time"`
	InconsistentData bool                  `json:"inconsistent_data"`
	Results          map[string][][]string `json:"results"`
}

type ReturnRooms struct {
	EndTime           time.Time      `json:"end_time"`
	StartTime         time.Time      `json:"start_time"`
	ContextSize       int            `json:"context_size"`
	InconsistentRooms []string       `json:"inconsistent_rooms"`
	Rooms             map[string]int `json:"rooms"`
}

type ReturnHistoricRooms struct {
	FirstTime time.Time                    `json:"start_time"`
	LastTime  time.Time                    `json:"last_time"`
	Rooms     map[string]map[time.Time]int `json:"rooms"`
}

type MacMetadata struct {
	AverageSignalStrength  float64   `json:"average_signal_strength"`
	DetectionCount         int       `json:"detection_count"`
	TypeCount              [3]int    `json:"type_count"`
	SSIDProbes             []string  `json:"ssid_probes"`
	HTCapabilities         *string   `json:"ht_capabilities"`
	HTExtendedCapabilities *string   `json:"ht_extended_capabilities"`
	SupportedRates         []float64 `json:"supported_rates"`
	Tags                   []int     `json:"tags"`
}

type ConfigResponse struct {
	SecondsPerWindow int `json:"seconds_per_window"`
}

type PersonalMacsUpload struct {
	DeviceID     string   `json:"device_id"`
	PersonalMacs []string `json:"personal_macs"`
}

type PersonalMacsDB struct {
	Mac      string `gorm:"primary_key" json:"mac"`
	Metadata string
}

func (PersonalMacsDB) TableName() string {
	return "personal_macs"
}

type RoomsDB struct {
	DeviceID string `json:"device_id"`
	RoomID   string `json:"room_id"`
}

func (RoomsDB) TableName() string {
	return "rooms"
}

type MacDigest struct {
	Mac                    string    `json:"mac"`
	AvgSignalStrength      float64   `json:"average_signal_strenght"`
	Manufacturer           *string   `json:"manufacturer"` // Manufacturer is nullable
	OuiID                  *string   `json:"oui_id"`
	TypeCount              [3]int    `json:"type_count"`
	PresenceRecord         []bool    `json:"presence_record"` // Last index in most recent
	SSIDProbes             []string  `json:"ssid_probes"`
	HTCapabilities         *string   `json:"ht_capabilities"`
	HTExtendedCapabilities *string   `json:"ht_extended_capabilities"`
	SupportedRates         []float64 `json:"supported_rates"`
	Tags                   []int     `json:"tags"`
}

func (m MacDigest) DistanceTo(other dbscan.Point) float64 {
	signalDifference := math.Abs(m.AvgSignalStrength - other.(MacDigest).AvgSignalStrength)
	signalDistance := 0.0
	if signalDifference > 1500 {
		signalDistance = 1
	} else {
		signalDistance = signalDifference / 1500.0
	}

	// TODO: check if correctly done
	typeDistance := 0.0
	for idx, v := range m.TypeCount {
		v1 := v
		v2 := other.(MacDigest).TypeCount[idx]
		diff := float64(v1 - v2)
		total := float64(v1 + v2)
		if diff != 0 {
			typeDistance = typeDistance + (math.Abs(diff)/total)/3.0
		}
	}

	presenceDistance := 0.0
	for idx, v := range m.PresenceRecord {
		if v != other.(MacDigest).PresenceRecord[idx] {
			presenceDistance = presenceDistance + 1.0/float64(len(m.PresenceRecord))
		}
	}

	distance := (2*signalDistance + typeDistance + presenceDistance) / 4.0
	//fmt.Printf("distance: %v\n", distance)

	return distance
}

func (m MacDigest) Name() string {
	return m.Mac
}

func GetStartEndTime() (time.Time, time.Time) {

	// Determine the start time and end time
	// The end time should be the current time with seconds truncated minus one
	now := time.Now()
	endTime := now.Truncate(60 * time.Second).Add(0 * time.Minute)

	// The start time should be the end time minus the windows context
	startTime := endTime.Add(-15 * time.Minute)

	cet, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		fmt.Println("Error getting start and end time: ", err.Error())
	}

	return startTime.In(cet), endTime.In(cet)

}

func GetLastTime() time.Time {

	now := time.Now()
	t := now.Truncate(60 * time.Second).Add(0 * time.Minute).Add(-1 * time.Minute)

	fmt.Printf("Generated time: %v\n", t)

	cet, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		fmt.Println("Error getting start and end time: ", err.Error())
	}

	return t.In(cet)
}
