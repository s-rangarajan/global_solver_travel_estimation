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
	ctx := context.TODO()

	sess := session.MustSession(session.NewSession())
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
	speedLookuper := NewSharedSpeedLookuper()
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				updatedSpeedMap, err := downloadSpeedMap(ctx, speedMapS3Key, speedMapS3Bucket, s3Downloader)
				if err != nil {
					log.Println(err.Error())
				}
				speedLookuper.UpdateSpeedMap(updatedSpeedMap)
			}
		}
	}()
	haversineEstimator := NewHaversineTravelEstimator(speedLookuper)

	http.HandleFunc("/compute_travel_estimates", func(w http.ResponseWriter, r *http.Request) {
		ComputeTravelEstimatesWithContext(ctx, haversineEstimator, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
