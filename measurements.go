package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Measurement struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Moisture    float64   `json:"moisture"`
}

type MeasurementInput struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Moisture    float64 `json:"moisture"`
}

type AverageMeasurement struct {
	AvgTemperature float64 `json:"avg_temperature"`
	AvgHumidity    float64 `json:"avg_humidity"`
	AvgMoisture    float64 `json:"avg_moisture"`
	Count          int     `json:"count"`
}

func HandlePostMeasurement(w http.ResponseWriter, r *http.Request) {
	var input MeasurementInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"INSERT INTO measurements (temperature, humidity, moisture) VALUES (?, ?, ?)",
		input.Temperature, input.Humidity, input.Moisture,
	)
	if err != nil {
		http.Error(w, "Failed to insert measurement", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
		"message": "Measurement recorded successfully",
	})
}

func HandleGetLatestMeasurements(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT id, timestamp, temperature, humidity, moisture FROM measurements ORDER BY timestamp DESC LIMIT 5",
	)
	if err != nil {
		http.Error(w, "Failed to retrieve measurements", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var measurements []Measurement
	for rows.Next() {
		var m Measurement
		err := rows.Scan(&m.ID, &m.Timestamp, &m.Temperature, &m.Humidity, &m.Moisture)
		if err != nil {
			continue
		}
		measurements = append(measurements, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(measurements)
}

func HandleGetAverageMeasurements(w http.ResponseWriter, r *http.Request) {
	var avg AverageMeasurement
	
	err := db.QueryRow(
		`SELECT 
			COALESCE(AVG(temperature), 0) as avg_temp,
			COALESCE(AVG(humidity), 0) as avg_hum,
			COALESCE(AVG(moisture), 0) as avg_moist,
			COUNT(*) as count
		FROM (
			SELECT temperature, humidity, moisture 
			FROM measurements 
			ORDER BY timestamp DESC 
			LIMIT 50
		)`,
	).Scan(&avg.AvgTemperature, &avg.AvgHumidity, &avg.AvgMoisture, &avg.Count)
	
	if err != nil {
		fmt.Printf("Error getting averages: %v\n", err)
		http.Error(w, "Failed to calculate averages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(avg)
}