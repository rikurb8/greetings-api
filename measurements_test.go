package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func setupTestDB(t *testing.T) func() {
	// Create a unique test database for each test
	tempDB := fmt.Sprintf("./test_%s_%d.db", t.Name(), time.Now().UnixNano())
	
	// Save original db
	originalDB := db
	
	// Initialize new test database
	err := InitTestDatabase(tempDB)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	
	// Return cleanup function
	return func() {
		if db != nil {
			db.Close()
		}
		db = originalDB
		os.Remove(tempDB)
	}
}

func InitTestDatabase(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS measurements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		temperature REAL,
		humidity REAL,
		moisture REAL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating measurements table: %w", err)
	}

	return nil
}

func TestHandlePostMeasurement(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create test measurement
	measurement := MeasurementInput{
		Temperature: 25.5,
		Humidity:    65.0,
		Moisture:    45.5,
	}

	body, _ := json.Marshal(measurement)
	req, err := http.NewRequest("POST", "/measurements", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlePostMeasurement)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NilError(t, err)
	assert.Equal(t, "Measurement recorded successfully", response["message"])
	assert.Assert(t, response["id"] != nil)
}

func TestHandleGetLatestMeasurements(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Insert test measurements
	for i := 0; i < 10; i++ {
		_, err := db.Exec(
			"INSERT INTO measurements (temperature, humidity, moisture) VALUES (?, ?, ?)",
			20.0+float64(i), 60.0+float64(i), 40.0+float64(i),
		)
		assert.NilError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	req, err := http.NewRequest("GET", "/measurements/latest", nil)
	assert.NilError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleGetLatestMeasurements)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var measurements []Measurement
	err = json.Unmarshal(rr.Body.Bytes(), &measurements)
	assert.NilError(t, err)
	assert.Equal(t, 5, len(measurements))

	// Verify that measurements are in descending order
	assert.Equal(t, 29.0, measurements[0].Temperature)
	assert.Equal(t, 28.0, measurements[1].Temperature)
}

func TestHandleGetAverageMeasurements(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Insert test measurements with same values
	for i := 0; i < 5; i++ {
		_, err := db.Exec(
			"INSERT INTO measurements (temperature, humidity, moisture) VALUES (?, ?, ?)",
			20.0, 60.0, 40.0,
		)
		assert.NilError(t, err)
	}

	req, err := http.NewRequest("GET", "/measurements/average", nil)
	assert.NilError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleGetAverageMeasurements)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var avg AverageMeasurement
	err = json.Unmarshal(rr.Body.Bytes(), &avg)
	assert.NilError(t, err)
	assert.Equal(t, 20.0, avg.AvgTemperature)
	assert.Equal(t, 60.0, avg.AvgHumidity)
	assert.Equal(t, 40.0, avg.AvgMoisture)
	assert.Equal(t, 5, avg.Count)
}

func TestHandlePostMeasurementInvalidJSON(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	req, err := http.NewRequest("POST", "/measurements", bytes.NewBufferString("invalid json"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlePostMeasurement)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Invalid request body\n", rr.Body.String())
}

func TestHandleGetAverageMeasurementsEmpty(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	req, err := http.NewRequest("GET", "/measurements/average", nil)
	assert.NilError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleGetAverageMeasurements)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var avg AverageMeasurement
	err = json.Unmarshal(rr.Body.Bytes(), &avg)
	assert.NilError(t, err)
	assert.Equal(t, 0.0, avg.AvgTemperature)
	assert.Equal(t, 0.0, avg.AvgHumidity)
	assert.Equal(t, 0.0, avg.AvgMoisture)
	assert.Equal(t, 0, avg.Count)
}