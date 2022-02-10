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

	for _, r := range uploadedData.ProbeRequest {

		// Insert data into database
		sql := `
		INSERT INTO data (device_id, station_bssid, ap_ssid, intent, time, power)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := db.Exec(sql, uploadedData.DeviceID, r.StationBssid, r.ApSsid, r.Intent, r.Time, r.Power)
		CheckError(err)

	}

	log.Printf("Inserted %d records", len(uploadedData.ProbeRequest))

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
	DeviceID     string `json:"device_id"`
	ProbeRequest []struct {
		ApSsid       string `json:"ap_ssid"`
		Intent       string `json:"intent"`
		Power        int64  `json:"power"`
		StationBssid string `json:"station_bssid"`
		Time         int64  `json:"time"`
	} `json:"probe_request"`
}
