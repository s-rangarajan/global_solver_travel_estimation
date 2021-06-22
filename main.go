package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	//ctx, signalCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	//defer signalCancel()

	haversineEstimator := NewHaversineTravelEstimator()
	http.HandleFunc("/compute_travel_estimates", func(w http.ResponseWriter, r *http.Request) {
		ComputeTravelEstimatesWithContext(context.TODO(), haversineEstimator, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
