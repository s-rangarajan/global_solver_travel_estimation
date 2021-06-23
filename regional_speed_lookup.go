package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type regionalSpeedMap struct {
	sync.RWMutex
	Values       map[int]map[int]map[int]float64
	DownloadedAt time.Time
}

var speedMap *regionalSpeedMap = new(regionalSpeedMap)

func LookupSpeedWithContext(
	ctx context.Context,
	serviceRegionID int,
	dispatchTime time.Time,
	downloadSpeedMapWithContext func(context.Context) (map[int]map[int]map[int]float64, error),
) (float64, error) {
	if speedMap.Values == nil || speedMap.DownloadedAt.Sub(time.Now()) > 30*time.Minute {
		if err := func() error {
			speedMap.Lock()
			defer speedMap.Unlock()
			values, err := downloadSpeedMapWithContext(ctx)
			if err != nil {
				return fmt.Errorf("error downloading regional speed map: %w", err)
			}
			speedMap.Values = values
			speedMap.DownloadedAt = time.Now()

			return nil
		}(); err != nil {
			return 0, err
		}
	}

	speedMap.RLock()
	defer speedMap.RUnlock()

	dayOfWeek := int(dispatchTime.Weekday())
	h, m, _ := dispatchTime.Clock()
	timeOfDay := h*60 + m
	if _, ok := speedMap.Values[serviceRegionID]; ok {
		if timeSpeedMap, ok := speedMap.Values[serviceRegionID][dayOfWeek]; ok {
			speed := -1.0
			for time := range timeSpeedMap {
				if time > timeOfDay {
					break
				}
				speed = timeSpeedMap[time]
			}
			if speed != -1 {
				return speed, nil
			}
		}
	}

	return 0, fmt.Errorf("unable to find speed for given region/time")
}
