package main

import (
	"context"
	"math"
	"runtime"
	"sync"
)

// radius of the earth in miles.
const earthRadiusMi = 3958

// Distance calculates the shortest path between two coordinates on the surface
// of the Earth. This function returns two units of measure, the first is the
// distance in miles, the second is the distance in kilometers.
func HaversineDistance(p, q Coordinates) float64 {
	lat1 := p.Latitude * math.Pi / 180
	lon1 := p.Longitude * math.Pi / 180
	lat2 := q.Latitude * math.Pi / 180
	lon2 := q.Longitude * math.Pi / 180

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c * earthRadiusMi
}

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
	EstimateTravelWithContext(context.Context, ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error)
}

type HaversineTravelEstimator struct {
	SpeedLookuper
}

func NewHaversineTravelEstimator(speedLookuper SpeedLookuper) *HaversineTravelEstimator {
	return &HaversineTravelEstimator{speedLookuper}
}

func (h *HaversineTravelEstimator) EstimateTravelWithContext(ctx context.Context, travelEstimatesRequest ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error) {
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
			localizedTravelSpeed, err := h.LookupSpeedWithContext(ctx, travelEstimatesRequest.ServiceRegionID, travelEstimatesRequest.DispatchTime)
			if err != nil {
				//TODO: log error
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
