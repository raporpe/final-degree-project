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
		INSERT INTO data_frames (bssid, station_mac, subtype, time, power, station_mac_vendor)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := db.Exec(sql, d.BSSID, d.StationMAC, d.Subtype, d.Time, d.Power, d.StationMACVendor)
		CheckError(err)

	}

	for _, c := range uploadedData.ControlFrames {

		// Insert data into database
		sql := `
		INSERT INTO control_frames (bssid, station_mac, subtype, time, power, station_mac_vendor)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := db.Exec(sql, c.BSSID, c.StationMAC, c.Subtype, c.Time, c.Power, c.StationMACVendor)
		CheckError(err)

	}

	for _, m := range uploadedData.ManagementFrames {

		// Insert data into database
		sql := `
		INSERT INTO management_frames (addr1, addr2, addr3, addr4, time, subtype, power)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := db.Exec(sql, m.Addr1, m.Addr2, m.Addr3, m.Addr4, m.Time, m.Subtype, m.Power)
		CheckError(err)

	}

	log.Printf("Inserted %d probes and %d beacons and %d controls", len(uploadedData.ProbeRequestFrames),
		len(uploadedData.BeaconFrames), len(uploadedData.ControlFrames))

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
	ControlFrames      []ControlFrame      `json:"control_frames"`
	ManagementFrames   []ManagementFrame   `json:"management_frames"`
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
	Subtype          int64   `json:"subtype"`
	Time             int64   `json:"time"`
	Power            int64   `json:"power"`
	StationMACVendor *string `json:"station_mac_vendor"` // So that the string is nullable
}

type ControlFrame struct {
	BSSID            string  `json:"bssid"`
	StationMAC       string  `json:"station_mac"`
	Subtype          string  `json:"subtype"`
	Time             int64   `json:"time"`
	Power            int64   `json:"power"`
	StationMACVendor *string `json:"station_mac_vendor"` // So that the string is nullable
}

type ManagementFrame struct {
	Addr1   *string `json:"addr1"`
	Addr2   *string `json:"addr2"`
	Addr3   *string `json:"addr3"`
	Addr4   *string `json:"addr4"`
	Time    int64   `json:"time"`
	Subtype string  `json:"subtype"` // So that the string is nullable
	Power   int64   `json:"power"`
}
