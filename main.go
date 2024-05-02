package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func main() {
	http.HandleFunc("/machine_memory_usage", getMachineMemoryUsage)
	http.HandleFunc("/cpu_usage", getCpuUsage)
	http.HandleFunc("/storage_statistics", getStorageUsage)
	log.Fatal(http.ListenAndServe(":8010", nil))
}

func queryFunction(r *http.Request, w http.ResponseWriter, query string) model.Value {
	prometheusURL := "http://aescontroller-monitoring-o-prometheus.aescloud-engine.svc.cluster.local:9090"
	client, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	prometheusAPI := v1.NewAPI(client)
	result, _, _ := prometheusAPI.Query(r.Context(), query, time.Now())
	return result
}

func getStorageUsage(w http.ResponseWriter, r *http.Request) {
	totalStorageQuery := `sum(node_filesystem_size_bytes)`
	availableStorageQuery := `sum(node_filesystem_size_bytes) - sum(node_filesystem_free_bytes)`

	totalStorage := queryFunction(r, w, totalStorageQuery)
	availableStorage := queryFunction(r, w, availableStorageQuery)

	jsonResponse := map[string]float64{"totalStorage": float64(totalStorage.(model.Vector)[0].Value), "availableStorage": float64(availableStorage.(model.Vector)[0].Value)}
	jsonBytes, err := json.Marshal(jsonResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func getCpuUsage(w http.ResponseWriter, r *http.Request) {
	cpuUsagePercentageQuery := `avg(sum by (instance, cpu) (rate(node_cpu_seconds_total{mode!~"idle|iowait|steal"}[5m])))`
	cpuUsagePercentage := queryFunction(r, w, cpuUsagePercentageQuery)

	jsonResponse := map[string]float64{"cpuUsage": float64(cpuUsagePercentage.(model.Vector)[0].Value)}
	jsonBytes, err := json.Marshal(jsonResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func getMachineMemoryUsage(w http.ResponseWriter, r *http.Request) {
	totalMemoryQuery := "machine_memory_bytes"
	memoryUsedQuery := "sum(container_memory_usage_bytes)"

	totalMemroy := 0
	usedMemroy := 0

	totalMemoryUsesResult := queryFunction(r, w, totalMemoryQuery)
	memoryUsedResult := queryFunction(r, w, memoryUsedQuery)

	for _, val := range totalMemoryUsesResult.(model.Vector) {
		totalMemroy += int(val.Value)
	}

	for _, val := range memoryUsedResult.(model.Vector) {
		usedMemroy += int(val.Value)
	}

	jsonResponse := map[string]float64{"totalRam": float64(totalMemroy), "usedRam": float64(usedMemroy)}
	jsonBytes, err := json.Marshal(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
