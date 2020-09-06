package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/wcharczuk/go-chart"
)

type metric struct {
	mutex      sync.Mutex
	datapoints map[string]datapoint
}

type datapoint struct {
	Value   float64 `json:"value"`
	Count   int     `json:"count"`
	Average float64 `json:"average"`
}

type error struct {
	Error string `json:"error"`
}

var data map[string]*metric
var globalLock sync.Mutex

const persistenceInterval = 30 * time.Second
const dataDir = "data"

func getMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	w.Header().Set("Content-Type", "application/json")

	// Find the requested data
	globalLock.Lock()
	m, ok := data[metricName]
	globalLock.Unlock()
	if ok {
		m.mutex.Lock()
		response, err := json.Marshal(m.datapoints)
		if err != nil {
			panic(err)
		}
		m.mutex.Unlock()
		w.Write(response)
	} else {
		response, err := json.Marshal(error{Error: "Could not find metric"})
		if err != nil {
			panic(err)
		}
		w.Write(response)
	}
}

func getMetricChart(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	dimensionName := chi.URLParam(r, "dimensionName")

	// Find the requested data
	globalLock.Lock()
	m, ok := data[metricName]
	globalLock.Unlock()
	if ok {
		m.mutex.Lock()

		// Build up dataset
		XValues := make([]time.Time, 0)
		YValues := make([]float64, 0)
		// TODO - sort timeseries
		for k, v := range m.datapoints {
			t, err := time.Parse("2006-01-02 15:04", k)
			if err != nil {
				fmt.Println("Could not parse time value")
				continue
			}
			XValues = append(XValues, t)

			switch dimensionName {
			case "sum":
				YValues = append(YValues, v.Value)
			case "count":
				YValues = append(YValues, float64(v.Count))
			default:
				YValues = append(YValues, v.Average)
			}
		}

		// Generate chart
		graph := chart.Chart{
			XAxis: chart.XAxis{
				ValueFormatter: chart.TimeHourValueFormatter,
			},
			Series: []chart.Series{
				chart.TimeSeries{
					XValues: XValues,
					YValues: YValues,
				},
			},
		}

		m.mutex.Unlock()
		w.Header().Set("Content-Type", "image/png")
		graph.Render(chart.PNG, w)
	} else {
		w.Header().Set("Content-Type", "application/json")
		response, err := json.Marshal(error{Error: "Could not find metric"})
		if err != nil {
			panic(err)
		}
		w.Write(response)
	}
}

func sampleData() {
	for {
		metricName := "test"
		timeKey := time.Now().Format("2006-01-02 15:04")
		value := float64(rand.Int63n(100))

		globalLock.Lock()
		m, ok := data[metricName]
		globalLock.Unlock()

		if ok {
			m.mutex.Lock()
			if point, ok := m.datapoints[timeKey]; ok {
				point.Count++
				point.Value += value
				point.Average = point.Value / float64(point.Count)
				m.datapoints[timeKey] = point
			} else {
				point := datapoint{Count: 1, Value: value, Average: value}
				m.datapoints[timeKey] = point
			}
			m.mutex.Unlock()
		} else {
			globalLock.Lock()
			data[metricName] = &metric{
				datapoints: make(map[string]datapoint, 0),
			}
			globalLock.Unlock()
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func restore() {
	globalLock.Lock()
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		fmt.Println("Data directory could not be read, restore skipped")
		return
	}
	for _, f := range files {
		metricName := strings.Replace(f.Name(), ".json", "", 1)
		datapoints := make(map[string]datapoint, 0)
		file, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dataDir, f.Name()))
		if err != nil {
			fmt.Printf("Could not read %s/%s\n", dataDir, f.Name())
			continue
		}
		json.Unmarshal(file, &datapoints)
		data[metricName] = &metric{
			datapoints: datapoints,
		}
	}
	globalLock.Unlock()
}

func persist() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0700)
	}
	for {
		globalLock.Lock()
		for metricName, m := range data {
			m.mutex.Lock()
			file, err := os.OpenFile(
				fmt.Sprintf("%s/%s.json", dataDir, metricName),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			if err != nil {
				fmt.Printf("Could not write %s to disk\n", metricName)
				continue
			}
			encoder := json.NewEncoder(file)
			encoder.Encode(m.datapoints)
			file.Close()
			if debugMode() {
				fmt.Printf("Persisted %s to disk\n", metricName)
			}
			m.mutex.Unlock()
		}
		globalLock.Unlock()
		time.Sleep(persistenceInterval)
	}
}

func debugMode() bool {
	return os.Getenv("DEBUG") != ""
}

func main() {
	// Set up router
	r := chi.NewRouter()
	r.Get("/metric/{metricName:[a-z-]+}", getMetric)
	r.Get("/metric/{metricName:[a-z-]+}/{dimensionName:[a-z-]+}.png", getMetricChart)
	// TODO - write API

	// Set up global data store
	data = make(map[string]*metric)
	restore()

	// TODO - LRU goroutine

	// Peristence goroutine
	go persist()

	// Spawn sample data goroutine
	go sampleData()

	// Start server
	port := "8080"
	fmt.Printf("Counter server started on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		panic(err)
	}
}
