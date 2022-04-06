package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

	// Read the query param start_time
	startTime, err := time.Parse(time.RFC3339, r.URL.Query().Get("start_time"))
	if err != nil {
		log.Println("Invalid start_time!: " + err.Error())
		w.Write([]byte("Invalid start_time"))
		return
	}

	// Read the query param end_time
	endTime, err := time.Parse(time.RFC3339, r.URL.Query().Get("end_time"))
	if err != nil {
		log.Println("Invalid end_time!")
		w.Write([]byte("Invalid end_time"))
		return
	}

	// Read the query param device_id
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		w.Write([]byte("device_id required"))
		return
	}

	result := GetDigestedMacs(deviceID, startTime, endTime)
	w.Write([]byte(result))

}

func GetDigestedMacs(deviceID string, startTime time.Time, endTime time.Time) string {

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
	jsonReturn, err := json.Marshal(&struct {
		NumberOfWindows   int                  `json:"number_of_windows"`
		WindowsStartTimes []time.Time          `json:"windows_start_times"`
		StartTime         time.Time            `json:"start_time"`
		EndTime           time.Time            `json:"end_time"`
		InconsistentData  bool                 `json:"inconsistent_data"`
		InconsistentTimes []time.Time          `json:"inconsistent_times"`
		Digest            map[string]MacDigest `json:"digest"`
	}{
		NumberOfWindows:   expectedWindowsBetween,
		WindowsStartTimes: expectedStartTimes,
		StartTime:         startTime,
		EndTime:           timePlusWindow(startTime, expectedWindowsBetween, windowSizeSeconds),
		Digest:            digestedMacs,
		InconsistentData:  inconsistentData,
		InconsistentTimes: inconsistentTimes,
	})
	if err != nil {
		log.Println("There was an error trying to marshall the final digested macs struct!")
		return ""
	}

	return string(jsonReturn)

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

type ReturnDetectedMacs struct {
	DeviceID         string                 `json:"device_id"`
	DetectedMacs     map[string]MacMetadata `json:"detected_macs"`
	SecondsPerWindow int                    `json:"seconds_per_window"`
	EndTime          time.Time              `json:"end_time"`
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

type MacDigest struct {
	AvgSignalStrength float64 `json:"average_signal_strenght"`
	Manufacturer      *string `json:"manufacturer"` // Manufacturer is nullable
	OuiID             string  `json:"oui_id"`
	TypeCount         [3]int  `json:"type_count"`
	PresenceRecord    []bool  `json:"presence_record"`
}
