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
	fmt.Printf("Serving configuration")
	configResponse := ConfigResponse{
		SecondsPerWindow: 60,
	}
	byteJson, err := json.Marshal(configResponse)
	CheckError(err)
	w.Write(byteJson)
}

func DigestedMacsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Serving mac digest")

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
	var deviceWindows []DetectedMacDB
	gormDB.Where("device_id = ? and start_time >= ? and start_time < ? ", deviceID, startTime, endTime).Find(&deviceWindows)

	// Calculate how many windows between
	secondsBetween := endTime.Sub(startTime).Seconds()
	windowsBetween := int(secondsBetween) / windowSizeSeconds

	// Check that no windows are missing
	if len(deviceWindows) < windowsBetween {
		log.Println("Window mismatch!!")
		fmt.Printf("len(deviceWindows): %v\n", len(deviceWindows))
		fmt.Printf("windowsBetween: %v\n", windowsBetween)
	}

	digestedMacs := make(map[string]MacDigest)

	for _, window := range deviceWindows {
		currentWindowNumber := int(window.StartTime.Sub(startTime).Seconds()) / windowSizeSeconds

		// Generate struct back from db
		var macMetadata map[string]MacMetadata
		err := json.Unmarshal([]byte(window.DetectedMacs), &macMetadata)
		if err != nil {
			fmt.Printf("Possible data corruption in db for window id %v ", window.ID)
			// Skip this window since information is not reliable
			continue
		}

		// Iterate over all the detected mac addresses
		for mac, data := range macMetadata {
			// If the mac exists in digestedMacsToReturn
			if m, exists := digestedMacs[mac]; exists {
				c := howManyTrue(m.PresenceRecord)
				m.AvgSignalStrength = int(math.Round(float64(digestedMacs[mac].AvgSignalStrength*c+data.AverageSignalStrength) / float64(c+1)))
				m.PresenceRecord[currentWindowNumber] = true

				// Assign new struct
				digestedMacs[mac] = m
			} else {
				// If the mac does no exist, create struct
				m := MacDigest{
					AvgSignalStrength: data.AverageSignalStrength,
					PresenceRecord:    make([]bool, windowsBetween),
				}
				// Set the presence record to true
				m.PresenceRecord[currentWindowNumber] = true

				// Assign new struct
				digestedMacs[mac] = m
			}

		}

	}

	// Return the digested macs
	jsonReturn, err := json.Marshal(digestedMacs)
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
	CheckError(err)

	var state UploadDetectedMacs

	err = json.Unmarshal(body, &state)
	CheckError(err)

	go StoreDetectedMacs(state)
}

func StoreDetectedMacs(upload UploadDetectedMacs) {
	fmt.Println("Storing state from " + upload.DeviceID)

	// Store the list of detected macs in the DB

	// Generate UUID
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Print("Error generating UUID for inserting the detected macs! Skip this insertion.")
		return
	}

	detectedMacs, err := json.Marshal(upload.DetectedMacs)
	if err != nil {
		log.Print("Error remarshalling the detected macs! Can this even happen?")
		return
	}

	gormDB.Create(&DetectedMacDB{
		ID:               uuid,
		DeviceID:         upload.DeviceID,
		SecondsPerWindow: upload.SecondsPerWindow,
		StartTime:        time.Unix(int64(upload.StartTime), 0),
		EndTime:          time.Unix(int64(upload.EndTime), 0),
		DetectedMacs:     string(detectedMacs),
	})

}

func GetVendor(mac string) *string {
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

type ReturnDetectedMacs struct {
	DeviceID         string                 `json:"device_id"`
	DetectedMacs     map[string]MacMetadata `json:"detected_macs"`
	SecondsPerWindow int                    `json:"seconds_per_window"`
	EndTime          time.Time              `json:"end_time"`
}

type MacMetadata struct {
	AverageSignalStrength int    `json:"average_signal_strength"`
	DetectionCount        int    `json:"detection_count"`
	Signature             string `json:"signature"`
	Typecount             []int  `json:"type_count"`
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

type MacDigest struct {
	AvgSignalStrength int    `json:"average_signal_strenght"`
	PresenceRecord    []bool `json:"presence_record"`
}
