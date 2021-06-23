package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	speedMapS3KeyEnvVar    = "REGIONAL_SPEED_MAP_KEY"
	speedMapS3BucketEnvVar = "REGION_SPEED_MAP_BUCKET_KEY"
)

func main() {
	//ctx, signalCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	//defer signalCancel()

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("cannot init session")
	}
	s3Downloader := s3manager.NewDownloader(sess)
	speedMapS3Key := os.Getenv(speedMapS3KeyEnvVar)
	if speedMapS3Key == "" {
		log.Fatalf("no key")
	}
	speedMapS3Bucket := os.Getenv(speedMapS3BucketEnvVar)
	if speedMapS3Key == "" {
		log.Fatalf("no bucket")
	}
	speedLookuper := func(ctx context.Context, serviceRegionID int, dispatchTime time.Time) (float64, error) {
		return LookupSpeedWithContext(
			ctx,
			serviceRegionID,
			dispatchTime,
			func(ctx context.Context) (map[int]map[int]map[int]float64, error) {
				return downloadSpeedMap(ctx, speedMapS3Key, speedMapS3Bucket, s3Downloader)
			},
		)
	}

	haversineEstimator := NewHaversineTravelEstimator(speedLookuper)
	http.HandleFunc("/compute_travel_estimates", func(w http.ResponseWriter, r *http.Request) {
		ComputeTravelEstimatesWithContext(context.TODO(), haversineEstimator, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
