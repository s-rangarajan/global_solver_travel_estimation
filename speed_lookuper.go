package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SpeedLookuper interface {
	LookupSpeedWithContext(context.Context, int, time.Time) (float64, error)
}

type sharedSpeedLookuper struct {
	sync.RWMutex
	speedMap map[int]map[int]map[int]float64
}

func NewSharedSpeedLookuper() *sharedSpeedLookuper {
	return &sharedSpeedLookuper{speedMap: make(map[int]map[int]map[int]float64)}
}

func (s *sharedSpeedLookuper) UpdateSpeedMap(speedMap map[int]map[int]map[int]float64) {
	s.Lock()
	defer s.Unlock()

	s.speedMap = speedMap
}

func (s *sharedSpeedLookuper) LookupSpeedWithContext(ctx context.Context, serviceRegionID int, dispatchTime time.Time) (float64, error) {
	s.RLock()
	defer s.RUnlock()

	dayOfWeek := int(dispatchTime.Weekday())
	h, m, _ := dispatchTime.Clock()
	timeOfDay := h*60 + m
	if _, ok := s.speedMap[serviceRegionID]; ok {
		if timeSpeedMap, ok := s.speedMap[serviceRegionID][dayOfWeek]; ok {
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
