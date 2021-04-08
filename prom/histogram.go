package prom

import (
	"fmt"

	"github.com/CXinsect/gocode/config"
)

type histogramWithLabels struct {
	labels  []string
	buckets map[float64]uint64
}

func transformHistogram(buckets map[float64]uint64, histogram config.Histogram) (transformed map[float64]uint64, count uint64, sum float64, err error) {
	if histogram.BucketType == config.HistogramBucketFixed {
		return transformFixed(buckets, histogram)
	} else {
		return transformDynamic(buckets, histogram)
	}
}

type generateFunc func(bucket float64) float64

func generateFinallyKeyWithMultiplier(histogram config.Histogram) (generateFunc, error) {
	mul := histogram.BucketMultiplier
	if mul == 0 {
		mul = 1
	}
	switch histogram.BucketType {
	case config.HistogramBucketLinear:
		return func(bucket float64) float64 {
			return mul * bucket
		}, nil
	case config.HistogramBucketFixed:
		return func(bucket float64) float64 {
			return mul * bucket
		}, nil
	default:
		return nil, fmt.Errorf("unknow type of histogram %q", histogram.BucketType)
	}
}

func transformFixed(bucket map[float64]uint64, histogram config.Histogram) (transformed map[float64]uint64, count uint64, sum float64, err error) {
	gFunc, err := generateFinallyKeyWithMultiplier(histogram)
	if err != nil {
		return nil, 0, 0, err
	}
	size := histogram.BucketMax - histogram.BucketMin
	if size == 0 {
		return nil, 0, 0, fmt.Errorf("histogram named %s bucket size is zero", histogram.Name)
	}
	transformed = make(map[float64]uint64, size)
	for i := float64(histogram.BucketMin); i < float64(histogram.BucketMax); i++ {
		if bucket[i] != 0 {
			count += bucket[i]
			transformed[gFunc(i)] = count
		} else {
			transformed[gFunc(i)] = 0
		}
	}

	mul := histogram.BucketMultiplier

	if mul == 0 {
		mul = 1
	}

	sum = float64(bucket[float64(histogram.BucketMax+1)]) * mul
	return
}
func transformDynamic(bucket map[float64]uint64, histogram config.Histogram) (transformed map[float64]uint64, count uint64, sum float64, err error) {
	gFunc, err := generateFinallyKeyWithMultiplier(histogram)

	if err != nil {
		return nil, 0, 0, err
	}
	size := len(histogram.BucketKeys)
	if size == 0 {
		return nil, 0, 0, fmt.Errorf("histogram named %s transformDynamic failed", histogram.Name)
	}

	transformed = make(map[float64]uint64, size)
	for i := 0; i < size; i++ {
		key := histogram.BucketKeys[i]

		if bucket[key] != 0 {
			count += bucket[key]
			transformed[gFunc(key)] = count
		} else {
			transformed[gFunc(key)] = 0
		}
	}

	mul := histogram.BucketMultiplier
	if mul == 0 {
		mul = 1
	}
	sum = float64(bucket[histogram.BucketKeys[size-1]+1]) * mul
	return
}
