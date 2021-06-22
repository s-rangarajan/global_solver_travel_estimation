package main

import (
	"context"
	"math"
	"runtime"
	"sync"
)

type TravelEstimateInput struct {
	FromLocation string
	ToLocation   string
	TravelSpeed  float64
}

type TravelEstimateResult struct {
	FromLocation string
	ToLocation   string
	Time         float64
	Distance     float64
}

type TravelEstimator interface {
	EstimateWithContext(context.Context, ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error)
}

type HaversineTravelEstimator struct {
	RegionalSpeedMap map[string]float64
}

func NewHaversineTravelEstimator() *HaversineTravelEstimator {
	return &HaversineTravelEstimator{}
}

func (h HaversineTravelEstimator) LookupSpeed(serviceRegionID int) (float64, bool) {
	return 0, false
}

func (h *HaversineTravelEstimator) EstimateWithContext(ctx context.Context, travelEstimatesRequest ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error) {
	responseChan := make(chan (ComputedTravelEstimates))
	go func() {
		var wg sync.WaitGroup
		computeChan := make(chan (TravelEstimateInput), 100000)
		resultsChan := make(chan (TravelEstimateResult), 100000)
		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for input := range computeChan {
					haversineDistance := HaversineDistance(
						travelEstimatesRequest.Locations[input.FromLocation],
						travelEstimatesRequest.Locations[input.ToLocation],
					)

					resultsChan <- TravelEstimateResult{
						FromLocation: input.FromLocation,
						ToLocation:   input.ToLocation,
						Time:         haversineDistance / input.TravelSpeed,
						Distance:     haversineDistance,
					}
				}
			}()
		}

		computedTravelEstimates := ComputedTravelEstimates{TravelEstimates: make(map[string]map[string]TravelEstimate, len(travelEstimatesRequest.Pairs))}
		for fromLocation := range travelEstimatesRequest.Pairs {
			computedTravelEstimates.TravelEstimates[fromLocation] = make(map[string]TravelEstimate, len(travelEstimatesRequest.Pairs[fromLocation]))
		}
		go func() {
			localizedTravelSpeed, ok := h.LookupSpeed(travelEstimatesRequest.ServiceRegionID)
			if !ok {
				localizedTravelSpeed = 15
			}
			for fromLocation := range travelEstimatesRequest.Pairs {
				for _, toLocation := range travelEstimatesRequest.Pairs[fromLocation] {
					computeChan <- TravelEstimateInput{
						FromLocation: fromLocation,
						ToLocation:   toLocation,
						TravelSpeed:  localizedTravelSpeed,
					}
				}
			}

			close(computeChan)
		}()

		go func() {
			wg.Wait()
			close(resultsChan)
		}()
		for result := range resultsChan {
			computedTravelEstimates.TravelEstimates[result.FromLocation][result.ToLocation] = TravelEstimate{
				Distance: math.Round(result.Distance*100) / 100,
				Time:     math.Round(result.Time*100) / 100,
			}
		}

		responseChan <- computedTravelEstimates
	}()

	select {
	case <-ctx.Done():
		return ComputedTravelEstimates{}, ctx.Err()
	case response := <-responseChan:
		return response, nil
	}
}
