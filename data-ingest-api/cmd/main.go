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
	r.HandleFunc("/v1/state", StateHandler)
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

	macDB, err = oui.OpenStaticFile("oui.txt")
	CheckError(err)

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

func StateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		response, err := json.Marshal(systemState)
		CheckError(err)

		w.Header().Add("Content-type", "application/json")
		w.Write(response)

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		CheckError(err)

		var state UploadedState

		err = json.Unmarshal(body, &state)
		CheckError(err)

		go StoreState(state)
	}
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

func StoreOcupationData(deviceID string, count int, t time.Time) {

	db := GetDB()

	sql := `
	INSERT INTO ocupation (device_id, count, time) values ($1, $2, $3)
	`

	_, err := db.Exec(sql, deviceID, count, t.Format(time.RFC3339))
	CheckError(err)

}

func StoreData(uploadedData *UploadJSON) {

	db := GetDB()

	for _, u := range uploadedData.ProbeRequestFrames {

		// Insert data into database
		sql := `
		INSERT INTO probe_request_frames (device_id, station_mac, intent, time, power, station_mac_vendor, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := db.Exec(sql, uploadedData.DeviceID, u.StationMAC,
			u.Intent, u.Time, u.Power, GetVendor(u.StationMAC), u.Frequency)
		CheckError(err)

	}

	// Insert probe responses
	for _, u := range uploadedData.ProbeReponseFrames {

		// Insert data into database
		sql := `
		INSERT INTO probe_response_frames (device_id, bssid, ssid, station_mac, station_mac_vendor, time, power, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		_, err := db.Exec(sql, uploadedData.DeviceID, u.BSSID, u.SSID,
			u.StationMAC, GetVendor(u.StationMAC), u.Time, u.Power, u.Frequency)
		CheckError(err)

	}

	for _, u := range uploadedData.BeaconFrames {

		// Insert data into database
		sql := `
		INSERT INTO beacon_frames (bssid, ssid, device_id, frequency)
		VALUES ($1, $2, $3, $4) ON CONFLICT (bssid) DO UPDATE SET frequency = $4
		`

		_, err := db.Exec(sql, u.BSSID, u.SSID, uploadedData.DeviceID, u.Frequency)
		CheckError(err)

	}

	for _, u := range uploadedData.DataFrames {

		// Insert data into database
		sql := `
		INSERT INTO data_frames (bssid, station_mac, subtype, time, power, station_mac_vendor, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := db.Exec(sql, u.BSSID, u.StationMAC, u.Subtype,
			u.Time, u.Power, GetVendor(u.StationMAC), u.Frequency)
		CheckError(err)

	}

	for _, u := range uploadedData.ControlFrames {

		// Insert data into database
		sql := `
		INSERT INTO control_frames (addr1, addr2, addr3, addr4, time, subtype, power, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		_, err := db.Exec(sql, u.Addr1, u.Addr2, u.Addr3, u.Addr4,
			u.Time, u.Subtype, u.Power, u.Frequency)
		CheckError(err)

	}

	for _, u := range uploadedData.ManagementFrames {

		// Insert data into database
		sql := `
		INSERT INTO management_frames (addr1, addr2, addr3, addr4, time, subtype, power, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		_, err := db.Exec(sql, u.Addr1, u.Addr2, u.Addr3, u.Addr4,
			u.Time, u.Subtype, u.Power, u.Frequency)

		CheckError(err)

	}

	probeRequests := len(uploadedData.ProbeRequestFrames)
	probeResponses := len(uploadedData.ProbeReponseFrames)
	beacons := len(uploadedData.BeaconFrames)
	controls := len(uploadedData.ControlFrames)
	datas := len(uploadedData.DataFrames)
	managements := len(uploadedData.ManagementFrames)
	total := probeRequests + probeResponses + beacons + controls + datas + managements

	log.Printf("Inserted \n %d probe request frames \n %d probe response frames \n %d beacon frames\n %d control frames\n %d data frames\n %d management frames \nTotal: %d",
		probeResponses, probeRequests,
		beacons, controls,
		datas, managements,
		total)

}

func GetDB() *sql.DB {

	if db == nil {

		// Stablish connection
		psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "tfg-server.raporpe.dev", 5432, "postgres", "raulportugues", "tfg")

		database, err := sql.Open("postgres", psqlconn)

		db = database

		CheckError(err)

	}

	return db

}

func GetVendor(mac string) *string {
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

type UploadJSON struct {
	DeviceID           string               `json:"device_id"`
	ProbeRequestFrames []ProbeRequestFrame  `json:"probe_request_frames"`
	BeaconFrames       []BeaconFrame        `json:"beacon_frames"`
	DataFrames         []DataFrame          `json:"data_frames"`
	ControlFrames      []ControlFrame       `json:"control_frames"`
	ManagementFrames   []ManagementFrame    `json:"management_frames"`
	ProbeReponseFrames []ProbeResponseFrame `json:"probe_response_frames"`
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

// Frames

type ProbeRequestFrame struct {
	StationMAC string  `json:"station_mac"`
	Intent     *string `json:"intent"`
	Time       string  `json:"time"`
	Frequency  int64   `json:"frequency"`
	Power      int64   `json:"power"`
}

type ProbeResponseFrame struct {
	BSSID      string `json:"bssid"`
	SSID       string `json:"ssid"`
	StationMAC string `json:"station_mac"`
	Time       string `json:"time"`
	Frequency  int64  `json:"frequency"`
	Power      int64  `json:"power"`
}

type BeaconFrame struct {
	SSID      string `json:"ssid"`
	BSSID     string `json:"bssid"`
	Frequency int64  `json:"frequency"`
}

type DataFrame struct {
	BSSID      string `json:"bssid"`
	StationMAC string `json:"station_mac"`
	Subtype    int64  `json:"subtype"`
	Time       string `json:"time"`
	Frequency  int64  `json:"frequency"`
	Power      int64  `json:"power"`
}

type ControlFrame struct {
	Addr1     *string `json:"addr1"`
	Addr2     *string `json:"addr2"`
	Addr3     *string `json:"addr3"`
	Addr4     *string `json:"addr4"`
	Time      string  `json:"time"`
	Subtype   string  `json:"subtype"` // So that the string is nullable
	Frequency int64   `json:"frequency"`
	Power     int64   `json:"power"`
}

type ManagementFrame struct {
	Addr1     *string `json:"addr1"`
	Addr2     *string `json:"addr2"`
	Addr3     *string `json:"addr3"`
	Addr4     *string `json:"addr4"`
	Time      string  `json:"time"`
	Subtype   string  `json:"subtype"` // So that the string is nullable
	Frequency int64   `json:"frequency"`
	Power     int64   `json:"power"`
}
