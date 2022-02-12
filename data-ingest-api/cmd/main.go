package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

var db *sql.DB

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/v1/upload", UploadHandler)

	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:2000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server...")

	log.Fatal(server.ListenAndServe())

}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var uploadedData UploadJSON

	err = json.Unmarshal(body, &uploadedData)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Received data from " + uploadedData.DeviceID)

	db := GetDB()

	for _, r := range uploadedData.ProbeRequestFrames {

		// Insert data into database
		sql := `
		INSERT INTO probe_request_frames (device_id, station_mac, intent, time, power, station_mac_vendor)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := db.Exec(sql, uploadedData.DeviceID, r.StationMAC, r.Intent, r.Time, r.Power, r.StationMACVendor)
		CheckError(err)

	}

	for _, b := range uploadedData.BeaconFrames {

		// Insert data into database
		sql := `
		INSERT INTO beacon_frames (bssid, ssid, device_id)
		VALUES ($1, $2, $3)
		`

		_, err := db.Exec(sql, b.BSSID, b.SSID, uploadedData.DeviceID)
		CheckError(err)

	}

	for _, d := range uploadedData.DataFrames {

		// Insert data into database
		sql := `
		INSERT INTO data_frames (bssid, station_mac, time, power, station_mac_vendor)
		VALUES ($1, $2, $3, $4, $5)
		`

		_, err := db.Exec(sql, d.BSSID, d.StationMAC, d.Time, d.Power, d.StationMACVendor)
		CheckError(err)

	}

	for _, a := range uploadedData.ActionFrames {

		// Insert data into database
		sql := `
		INSERT INTO action_frames (bssid, station_mac, subtype, time, power, station_mac_vendor)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := db.Exec(sql, a.BSSID, a.StationMAC, a.Subtype, a.Time, a.Power, a.StationMACVendor)
		CheckError(err)

	}

	log.Printf("Inserted %d probes and %d beacons", len(uploadedData.ProbeRequestFrames), len(uploadedData.BeaconFrames))

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

func CheckError(err error) {
	if err != nil {
		log.Println(err)
	}
}

type UploadJSON struct {
	DeviceID           string              `json:"device_id"`
	ProbeRequestFrames []ProbeRequestFrame `json:"probe_request_frames"`
	BeaconFrames       []BeaconFrame       `json:"beacon_frames"`
	DataFrames         []DataFrame         `json:"data_frames"`
	ActionFrames       []ActionFrame       `json:"action_frames"`
}

type ProbeRequestFrame struct {
	StationMAC       string  `json:"station_mac"`
	Intent           *string `json:"intent"`
	Time             int64   `json:"time"`
	Power            int64   `json:"power"`
	StationMACVendor *string `json:"station_mac_vendor"` // So that the string is nullable
}

type BeaconFrame struct {
	SSID  string `json:"ssid"`
	BSSID string `json:"bssid"`
}

type DataFrame struct {
	BSSID            string  `json:"bssid"`
	StationMAC       string  `json:"station_mac"`
	Time             int64   `json:"time"`
	Power            int64   `json:"power"`
	StationMACVendor *string `json:"station_mac_vendor"` // So that the string is nullable
}

type ActionFrame struct {
	BSSID            string  `json:"bssid"`
	StationMAC       string  `json:"station_mac"`
	Subtype          string  `json:"subtype"`
	Time             int64   `json:"time"`
	Power            int64   `json:"power"`
	StationMACVendor *string `json:"station_mac_vendor"` // So that the string is nullable
}
