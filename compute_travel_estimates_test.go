package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

/*
`
	{
		"locations": {
			"r1234": {
				"latitude": 1.1,
				"longitude": 1.1
			},
			"d1234": {
				"latitude": 1.1,
				"longitude": 1.1
			},
			"c1234": {
				"latitude": 1.1,
				"longitude": 1.1
			},
			"c5678": {
				"latitude": 1.2,
				"longitude": 1.2
			}
		},
		"pairs": {
			"r1234": ["c1234", "c5678"],
			"d1234": ["r1234"],
			"c1234": ["c5678"],
			"c5678": ["c1234"]
		}
	}
`
*/

func TestGenerate(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	restaurants := make(map[string]Coordinates, 100)
	drivers := make(map[string]Coordinates, 50)
	customers := make(map[string]Coordinates, 100)

	pairs := make(map[string][]string, 250)

	for i := 1; i <= 100; i++ {
		lats := randFloats(-90, 90, 1)
		longs := randFloats(-180, 180, 1)
		restaurants[fmt.Sprintf("r%d", i)] = Coordinates{lats[0], longs[0]}
		customers[fmt.Sprintf("c%d", i)] = Coordinates{lats[0], longs[0]}
	}

	for i := 1; i <= 50; i++ {
		lats := randFloats(-90, 90, 1)
		longs := randFloats(-180, 180, 1)
		drivers[fmt.Sprintf("d%d", i)] = Coordinates{lats[0], longs[0]}
	}

	for fromNode := range restaurants {
		pairs[fromNode] = make([]string, 0, 0)
		for toNode := range restaurants {
			if toNode != fromNode {
				pairs[fromNode] = append(pairs[fromNode], toNode)
			}
		}

		for toNode := range customers {
			pairs[fromNode] = append(pairs[fromNode], toNode)
		}
	}

	for fromNode := range drivers {
		pairs[fromNode] = make([]string, 0, 0)
		for toNode := range restaurants {
			pairs[fromNode] = append(pairs[fromNode], toNode)
		}
	}

	for fromNode := range customers {
		pairs[fromNode] = make([]string, 0, 0)
		for toNode := range customers {
			if toNode != fromNode {
				pairs[fromNode] = append(pairs[fromNode], toNode)
			}
		}
	}

	locations := make(map[string]Coordinates, 250)
	for restaurant := range restaurants {
		locations[restaurant] = restaurants[restaurant]
	}
	for driver := range drivers {
		locations[driver] = drivers[driver]
	}
	for customer := range customers {
		locations[customer] = customers[customer]
	}

	request := ComputeTravelEstimatesRequest{
		ServiceRegionID: 1,
		Locations:       locations,
		Pairs:           pairs,
	}

	j, _ := json.MarshalIndent(request, "", "	")
	os.WriteFile("request", j, 0644)
}

func randFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}
