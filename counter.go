package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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
		time.Sleep(200)
	}
}

func main() {
	// Set up router
	r := chi.NewRouter()
	r.Get("/metric/{metricName:[a-z-]+}", getMetric)
	r.Get("/metric/{metricName:[a-z-]+}/{dimensionName:[a-z-]+}.png", getMetricChart)
	// TODO - read API

	// Set up global data store
	data = make(map[string]*metric)

	// TODO - LRU goroutine

	// TODO - peristence goroutine

	// Spawn sample data goroutine
	go sampleData()

	// Start server
	port := "8080"
	fmt.Printf("Counter server started on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		panic(err)
	}
}
