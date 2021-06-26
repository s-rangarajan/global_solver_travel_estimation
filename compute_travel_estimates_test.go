package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockTravelEstimator struct {
	TestEstimateWithContext func(context.Context, ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error)
}

func (m *MockTravelEstimator) EstimateWithContext(ctx context.Context, request ComputeTravelEstimatesRequest) (ComputedTravelEstimates, error) {
	return m.TestEstimateWithContext(ctx, request)
}

func TestComputeTravelEstimatesWithContextReturnsErrorIfInvalidJSON(t *testing.T) {
	request, err := http.NewRequest("POST", "/compute_travel_estimates", bytes.NewBuffer([]byte("totally not JSON")))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	// can be nil because should not be invoked
	ComputeTravelEstimatesWithContext(context.Background(), &MockTravelEstimator{}, response, request)

	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
	var errorMessage struct {
		Message string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&errorMessage))
	require.Regexp(t, "unmarshaling request", errorMessage.Message)
}

func TestComputeTravelEstimatesWithContextReturnsErrorIfInvalidDataStructure(t *testing.T) {
	request, err := http.NewRequest("POST", "/compute_travel_estimates", bytes.NewBuffer([]byte(`"valid JSON"`)))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	// can be nil because should not be invoked
	ComputeTravelEstimatesWithContext(context.Background(), &MockTravelEstimator{}, response, request)

	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
	var errorMessage struct {
		Message string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&errorMessage))
	require.Regexp(t, "unmarshaling request", errorMessage.Message)
}

func TestComputeTravelEstimatesWithContextReturnsErrorIfInvalidData(t *testing.T) {
	request, err := http.NewRequest("POST", "/compute_travel_estimates", bytes.NewBuffer([]byte(`{}`)))
	require.NoError(t, err)

	response := httptest.NewRecorder()
	// can be nil because should not be invoked
	ComputeTravelEstimatesWithContext(context.Background(), &MockTravelEstimator{}, response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	var errorMessage struct {
		Message string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&errorMessage))
	require.Regexp(t, "invalid request", errorMessage.Message)
}
