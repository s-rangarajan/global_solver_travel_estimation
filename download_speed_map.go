package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

func downloadSpeedMap(ctx context.Context, speedMapKey, speedMapBucket string, downloader s3manageriface.DownloaderAPI) (map[int]map[int]map[int]float64, error) {
	buffer := aws.NewWriteAtBuffer(make([]byte, 0))

	_, err := downloader.DownloadWithContext(
		ctx,
		buffer,
		&s3.GetObjectInput{
			Key:    aws.String(speedMapKey),
			Bucket: aws.String(speedMapBucket),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error downloading speed map: %w", err)
	}

	speedMap := make(map[int]map[int]map[int]float64)
	csvReader := csv.NewReader(bytes.NewBuffer(buffer.Bytes()))
	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing csv file: %w", err)
		}
		if len(line) != 4 {
			return nil, fmt.Errorf("line in file has unexpected number of columns: %+v", line)
		}

		serviceRegion, err := strconv.ParseInt(line[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing region from line: %w", err)
		}
		if serviceRegion < 0 {
			return nil, fmt.Errorf("unexpected value for serviceRegion: %+v", serviceRegion)
		}
		dayOfWeek, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing dayOfWeek from line: %w", err)
		}
		if dayOfWeek < 0 || dayOfWeek > 6 {
			return nil, fmt.Errorf("unexpected value for dayOfWeek: %+v", dayOfWeek)
		}
		timeOfDay, err := strconv.ParseInt(line[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing timeOfDay from line: %w", err)
		}
		if timeOfDay < 0 || timeOfDay > 1440 {
			return nil, fmt.Errorf("unexpected value for timeOfDay: %+v", timeOfDay)
		}
		speed, err := strconv.ParseFloat(line[3], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing speed from line: %w", err)
		}
		if speed < 0 {
			return nil, fmt.Errorf("unexpected value for speed: %+v", speed)
		}

		if _, ok := speedMap[int(serviceRegion)]; !ok {
			speedMap[int(serviceRegion)] = make(map[int]map[int]float64)
		}
		if _, ok := speedMap[int(serviceRegion)][int(dayOfWeek)]; !ok {
			speedMap[int(serviceRegion)][int(dayOfWeek)] = make(map[int]float64)
		}
		speedMap[int(serviceRegion)][int(dayOfWeek)][int(timeOfDay)] = speed
	}

	return speedMap, nil
}
