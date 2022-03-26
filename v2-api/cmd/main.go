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

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/klauspost/oui"
)

// The mac vendor IEEE database
var macDB oui.StaticDB

// Database initialization
var gormDB *gorm.DB

func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "tfg-server.raporpe.dev", 5432, "postgres", "raulportugues", "tfg")

	var err error
	gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	gormDB.AutoMigrate(&DetectedMacsTable{})

	r := mux.NewRouter()
	r.HandleFunc("/v1/detected-macs", PostDetectedMacsHandler).Methods("POST")
	r.HandleFunc("/v1/detected-macs", GetDetectedMacsHandler).Methods("GET")
	r.HandleFunc("/v1/config", ConfigHandler)

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

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Serving configuration")
	configResponse := ConfigResponse{
		SecondsPerWindow: 60,
	}
	byteJson, err := json.Marshal(configResponse)
	CheckError(err)
	w.Write(byteJson)
}

func GetDetectedMacsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	startTime, err := time.Parse(time.RFC3339, mux.Vars(r)["start_time"])
	if err != nil {
		log.Println("Invalid start_time!")
		w.Write([]byte("Invalid start_time"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, mux.Vars(r)["start_time"])
	if err != nil {
		log.Println("Invalid end_time!")
		w.Write([]byte("Invalid end_time"))
		return
	}

	// Get data from the database
	var ret []DetectedMacsTable
	gormDB.Where(&DetectedMacsTable{StartTime: startTime, EndTime: endTime}).Find(&ret)

	jsonResponse, err := json.Marshal(ret)
	if err != nil {
		log.Println("Error generating json with data from the Detected Macs table!")
		w.Write([]byte("There was an error serializing request"))
	}
	w.Write(jsonResponse)

}

func PostDetectedMacsHandler(w http.ResponseWriter, r *http.Request) {
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

	gormDB.Create(&DetectedMacsTable{
		ID:           uuid,
		DeviceID:     upload.DeviceID,
		StartTime:    time.Unix(int64(upload.StartTime), 0),
		EndTime:      time.Unix(int64(upload.EndTime), 0),
		DetectedMacs: string(detectedMacs),
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
	DeviceID         string        `json:"device_id"`
	DetectedMacs     []MacMetadata `json:"detected_macs"`
	SecondsPerWindow int           `json:"seconds_per_window"`
	NumberOfWindows  int           `json:"number_of_windows"`
	StartTime        int           `json:"start_time"`
	EndTime          int           `json:"end_time"`
}

type MacMetadata struct {
	AverageSignalStrength int    `json:"average_signal_strength"`
	DetectionCount        int    `json:"detection_count"`
	Signature             string `json:"signature"`
	Typecount             []int  `json:"type_count"`
}

type ConfigResponse struct {
	SecondsPerWindow int `json:"secondsPerWindow"`
}

type DetectedMacsTable struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;"`
	DeviceID     string
	StartTime    time.Time
	EndTime      time.Time
	DetectedMacs string
}
