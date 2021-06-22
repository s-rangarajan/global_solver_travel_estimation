package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const computeTimeout = 3 * time.Second

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ComputeTravelEstimatesRequest struct {
	ServiceRegionID int                    `json:"service_region_id"`
	Locations       map[string]Coordinates `json:"locations"`
	Pairs           map[string][]string    `json:"pairs"`
}

type TravelEstimate struct {
	Distance float64 `json:"distance"`
	Time     float64 `json:time"`
}

type ComputedTravelEstimates struct {
	TravelEstimates map[string]map[string]TravelEstimate `json:"travel_estimates"`
}

func ComputeTravelEstimatesWithContext(ctx context.Context, estimator TravelEstimator, w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(ctx, computeTimeout)
	defer cancelFunc()

	var computeRequest ComputeTravelEstimatesRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&computeRequest); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	computedValues, err := estimator.EstimateWithContext(ctx, computeRequest)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(computedValues); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
