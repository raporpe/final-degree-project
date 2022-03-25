package main

import (
	"database/sql"
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

var db *sql.DB
var macDB oui.StaticDB

var systemState = make(map[string]map[time.Time]map[string]MacState)

// Database initialization
var gormDB *gorm.DB

func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "tfg-server.raporpe.dev", 5432, "postgres", "raulportugues", "tfg")

	var err error
	gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	gormDB.AutoMigrate(&SystemStateStore{})

	r := mux.NewRouter()
	r.HandleFunc("/v1/upload", UploadHandler)
	r.HandleFunc("/v1/state", PostStateHandler).Methods("POST")
	r.HandleFunc("/v1/state/{time}", GetStateHandler).Methods("GET")
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
	fmt.Printf("Serving config")
	configResponse := ConfigResponse{
		WindowTime: 60,
		WindowSize: 15,
	}
	byteJson, err := json.Marshal(configResponse)
	CheckError(err)
	w.Write(byteJson)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	CheckError(err)

	var uploadedData UploadJSON

	err = json.Unmarshal(body, &uploadedData)
	CheckError(err)

	log.Println("-------------------------------------------")
	log.Println("Received data from " + uploadedData.DeviceID)
	go StoreData(&uploadedData)

}

func PostStateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	CheckError(err)

	var state UploadedState

	err = json.Unmarshal(body, &state)
	CheckError(err)

	go StoreState(state)
}

func GetStateHandler(w http.ResponseWriter, r *http.Request) {
	requestedTime, err := time.Parse(time.RFC3339, mux.Vars(r)["time"])
	CheckError(err)

	s := make(map[string]map[time.Time]map[string]MacState)
	for device, date := range systemState {
		s[device] = make(map[time.Time]map[string]MacState)
		s[device][requestedTime] = date[requestedTime]
	}
	response, err := json.Marshal(s)
	CheckError(err)

	w.Header().Add("Content-type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Write(response)

}

func StoreState(uState UploadedState) {
	fmt.Println("Storing state from " + uState.DeviceID)

	deviceID := uState.DeviceID
	t := time.Unix(int64(uState.Time), 0)

	// If device is new
	if systemState[deviceID] == nil {
		systemState[deviceID] = make(map[time.Time]map[string]MacState)
	}

	// Time has not been previously registered
	if systemState[deviceID][t] == nil {
		systemState[deviceID][t] = make(map[string]MacState)
	}

	activeMacs := 0
	// Each iteration is the record of a single mac
	for _, s := range uState.MacStates {
		// Convert the string to the bitset state
		newRecord := NewRecord(s.Record)

		systemState[deviceID][t][s.Mac] = MacState{
			Record:         *newRecord,
			SignalStrength: int64(s.SignalStrength),
		}

		if newRecord.IsActive() {
			activeMacs++
		}

	}

	StoreOcupationData(uState.DeviceID, activeMacs, t)

	j, err := json.Marshal(systemState[deviceID][t])
	CheckError(err)

	// Store the state in the DB
	uuid, err := uuid.NewUUID()
	CheckError(err)
	gormDB.Create(&SystemStateStore{
		ID:          uuid,
		DeviceID:    uState.DeviceID,
		Time:        t,
		DeviceState: string(j),
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

type OcupationData struct {
	DeviceID string `json:"device_id"`
	Count    int64  `json:"count"`
}

type UploadedState struct {
	DeviceID         string             `json:"device_id"`
	MacStates        []UploadedMacState `json:"mac_states"`
	SecondsPerWindow int                `json:"seconds_per_window"`
	NumberOfWindows  int                `json:"number_of_windows"`
	Time             int                `json:"time"`
}

type UploadedMacState struct {
	Mac            string `json:"mac"`
	Record         string `json:"record"`
	SignalStrength int    `json:"signal_strength"`
}

type ConfigResponse struct {
	WindowTime int `json:"window_time"`
	WindowSize int `json:"window_size"`
}

type DeviceState map[int]MacState

type MacState struct {
	Record         Record `json:"record"`
	SignalStrength int64  `json:"signal_strength"`
}

type SystemStateStore struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	DeviceID    string
	Time        time.Time
	DeviceState string
}
