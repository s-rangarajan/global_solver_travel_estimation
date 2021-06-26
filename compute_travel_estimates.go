package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	DispatchTime    time.Time              `json:"dispatch_time"`
	Locations       map[string]Coordinates `json:"locations"`
	Pairs           map[string][]string    `json:"pairs"`
}

func (c ComputeTravelEstimatesRequest) Validate() error {
	if c.ServiceRegionID <= 0 {
		return fmt.Errorf("missing/invalid service region id")
	}
	if c.DispatchTime.IsZero() {
		return fmt.Errorf("missing dispatch time")
	}
	if c.Locations == nil {
		return fmt.Errorf("missing locations")
	}
	if c.Pairs == nil {
		return fmt.Errorf("missing pairs")
	}

	return nil
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

	w.Header().Set("Content-Type", "application/json")
	var computeRequest ComputeTravelEstimatesRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&computeRequest); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, fmt.Errorf("error unmarshaling request: %w", err).Error())))
		return
	}

	if err := computeRequest.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, fmt.Errorf("invalid request: %w", err).Error())))
		return
	}

	computedValues, err := estimator.EstimateTravelWithContext(ctx, computeRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(computedValues); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
