package main

import (
	"encoding/json"
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
		log.Println("There was an error trying to marshall the final digested macs struct!")
		return
	}

	w.Write([]byte(jsonResponse))

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

	clusteredMacs, err := GetClusteredMacs(roomID)
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

func GetClusteredMacs(roomID string) (map[string][][]string, error) {

	// Get the devices that are in the room
	var CaptureDevicesInRoom []CaptureDevicesDB
	gormDB.Where("room_id = ?", roomID).Find(&CaptureDevicesInRoom)

	// Get the start time and end time
	startTime, endTime := GetStartEndTime()

	// Create the return variable, a map from devices to clusters
	toReturn := make(map[string][][]string)

	// Iterate over every device in the room
	for _, device := range CaptureDevicesInRoom {
		deviceID := device.DeviceID
		digestedMacs := GetDigestedMacs(deviceID, startTime, endTime)

		// Exclude mac addresses that are not active
		//var activeMacs []string

		// Separate those macs that are real
		var realMacs []string
		for _, v := range digestedMacs.Digest {
			if v.Manufacturer != nil {
				realMacs = append(realMacs, v.Mac)
			}
		}

		// Convert from MacDigest type into dbscan.Point type
		// Only fake macs will go trough this process
		var points []dbscan.Point
		for _, v := range digestedMacs.Digest {
			if v.Manufacturer == nil {
				points = append(points, v)
			}
		}

		DBScanResult := dbscan.Cluster(0, 0.2, points)

		// Convert from points into strings (macs)
		var clusters [][]string

		for _, c := range DBScanResult {
			var cluster []string
			for _, p := range c {
				cluster = append(cluster, p.(MacDigest).Mac)
			}
			clusters = append(clusters, cluster)
		}

		fmt.Printf("DBScanResult: %v\n", len(DBScanResult))

		// Insert the real macs as unique clusters
		//for _, r := range realMacs {
		//	clusters = append(clusters, []string{r})
		//}

		fmt.Printf("starting macs: %v\n", len(digestedMacs.Digest))
		endingMacs := 0
		for _, i := range clusters {
			for range i {
				endingMacs += 1
			}
		}

		fmt.Printf("ending macs: %v\n", endingMacs)

		toReturn[deviceID] = clusters

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
				log.Printf("SKIPPING windo w.StartTime: %v\n", expectedTime)
			}
		}

		if skip {
			log.Printf("SKIPPING window.StartTime: %v\n", expectedTime)
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
		log.Printf("currentWindowNumber: %v\n", currentWindowNumber)

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

				// Assign the modified struct to the digested macs
				digestedMacs[mac] = m
			} else {
				// If the mac does no exist, create struct
				m := MacDigest{
					Mac:               mac,
					AvgSignalStrength: data.AverageSignalStrength,
					PresenceRecord:    make([]bool, expectedWindowsBetween),
					TypeCount:         data.TypeCount,
					Manufacturer:      GetMacVendor(mac),
					OuiID:             GetMacPrefix(mac),
				}
				// Set the presence record to true
				m.PresenceRecord[currentWindowNumber] = true

				// Assign new struct
				digestedMacs[mac] = m
			}

		}

	}

	// Return the digested macs
	return ReturnDigestedMacs{
		NumberOfWindows:   expectedWindowsBetween,
		WindowsStartTimes: expectedStartTimes,
		StartTime:         startTime,
		EndTime:           timePlusWindow(startTime, expectedWindowsBetween, windowSizeSeconds),
		Digest:            digestedMacs,
		InconsistentData:  inconsistentData,
		InconsistentTimes: inconsistentTimes,
	}

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
		log.Println("Error when unmarshalling received body")
		w.WriteHeader(500)
		return
	}

	log.Println("Storing detected macs from " + state.DeviceID)

	// Generate UUID
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Print("Error generating UUID for inserting the detected macs! Skip this insertion.")
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

func GetMacPrefix(mac string) string {
	if macDB == nil {
		var err error
		macDB, err = oui.OpenStaticFile("oui.txt")
		CheckError(err)
	}

	result, err := macDB.Query(mac)
	if err != nil {
		return ""
	} else {
		return result.Prefix.String()
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
}

type MacMetadata struct {
	AverageSignalStrength float64 `json:"average_signal_strength"`
	DetectionCount        int     `json:"detection_count"`
	Signature             string  `json:"signature"`
	TypeCount             [3]int  `json:"type_count"`
}

type ConfigResponse struct {
	SecondsPerWindow int `json:"seconds_per_window"`
}

type PersonalMacsUpload struct {
	DeviceID     string   `json:"device_id"`
	PersonalMacs []string `json:"personal_macs"`
}

type PersonalMacsDB struct {
	Mac string `gorm:"primary_key" json:"mac"`
}

func (PersonalMacsDB) TableName() string {
	return "personal_macs"
}

type CaptureDevicesDB struct {
	DeviceID string `json:"device_id"`
	RoomID   string `json:"room_id"`
}

func (CaptureDevicesDB) TableName() string {
	return "capture_devices"
}

type MacDigest struct {
	Mac               string  `json:"mac"`
	AvgSignalStrength float64 `json:"average_signal_strenght"`
	Manufacturer      *string `json:"manufacturer"` // Manufacturer is nullable
	OuiID             string  `json:"oui_id"`
	TypeCount         [3]int  `json:"type_count"`
	PresenceRecord    []bool  `json:"presence_record"`
}

func (m MacDigest) DistanceTo(other dbscan.Point) float64 {
	signalDifference := math.Abs(m.AvgSignalStrength - other.(MacDigest).AvgSignalStrength)
	signalDistance := 0.0
	if signalDifference > 1500 {
		signalDistance = 1
	} else {
		signalDistance = signalDifference / 1500.0
	}

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

	fmt.Printf("typeDistance: %v\n", typeDistance)

	presenceDistance := 0.0
	for idx, v := range m.PresenceRecord {
		if v != other.(MacDigest).PresenceRecord[idx] {
			presenceDistance = presenceDistance + 1.0/float64(len(m.PresenceRecord))
		}
	}

	fmt.Printf("presenceDistance: %v\n", presenceDistance)

	distance := (2*signalDistance + typeDistance + presenceDistance) / 4.0
	fmt.Printf("distance: %v\n", distance)

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
