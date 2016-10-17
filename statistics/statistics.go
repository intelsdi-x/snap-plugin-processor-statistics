/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statistics

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/montanaflynn/stats"
)

const (
	Name    = "statistics"
	Version = 2
)

type Plugin struct {
	buffer        map[string][]interface{}
	bufferMaxSize int
	bufferCurSize int
	bufferIndex   int
}

// New() returns a new instance of this
func New() *Plugin {
	buffer := make(map[string][]interface{})
	p := &Plugin{buffer: buffer,
		bufferMaxSize: 100,
		bufferCurSize: 0,
		bufferIndex:   0}
	return p
}

// calculateStats calaculates the descriptive statistics for buff
func (p *Plugin) calculateStats(buff interface{}, startTime time.Time, stopTime time.Time, namespace string, unit string) ([]plugin.Metric, error) {
	var result []plugin.Metric
	var buffer []float64
	var err error
	var val float64
	var modeVal []float64
	tags := map[string]string{
		"startTime": startTime.String(),
		"stopTime":  stopTime.String(),
	}
	time := time.Now()

	//Need to change so it ranges over the current size of the buffer and not the capacity
	for _, val := range buff.([]interface{}) {
		switch v := val.(type) {
		default:
			st := fmt.Sprintf("Unknown data received in calculateStats(): Type %T", v)
			return nil, errors.New(st)
		case int:
			buffer = append(buffer, float64(val.(int)))
		case int32:
			buffer = append(buffer, float64(val.(int32)))
		case int64:
			buffer = append(buffer, float64(val.(int64)))
		case float64:
			buffer = append(buffer, val.(float64))
		case float32:
			buffer = append(buffer, float64(val.(float32)))
		case uint64:
			buffer = append(buffer, float64(val.(uint64)))
		case uint32:
			buffer = append(buffer, float64(val.(uint32)))
		}
	}

	statList := [...]string{"Count", "Mean", "Median", "Standard Deviation", "Variance", "95%-ile", "99%-ile", "2%-ile", "9%-ile", "25%-ile", "75%-ile", "91%-ile", "98%-ile", "Minimum", "Maximum", "Range", "Mode", "Kurtosis", "Skewness", "Sum", "Trimean", "Quartile_Range"}
	mean, meanErr := stats.Mean(buffer)
	stdev, stdevErr := stats.StandardDeviation(buffer)
	min, minErr := stats.Min(buffer)
	max, maxErr := stats.Max(buffer)
	median, medErr := stats.Median(buffer)

	for _, stat := range statList {
		switch stat {
		case "Count":
			val = float64(len(buffer))
		case "Mean":
			val, err = mean, meanErr
		case "Median":
			val, err = median, medErr
		case "Standard Deviation":
			val, err = stdev, stdevErr
		case "Variance":
			val, err = stats.Variance(buffer)
		case "95%-ile":
			val, err = stats.PercentileNearestRank(buffer, 95)
		case "99%-ile":
			val, err = stats.PercentileNearestRank(buffer, 99)
		case "2%-ile":
			val, err = stats.PercentileNearestRank(buffer, 2)
		case "9%-ile":
			val, err = stats.PercentileNearestRank(buffer, 9)
		case "25%-ile":
			val, err = stats.PercentileNearestRank(buffer, 25)
		case "75%-ile":
			val, err = stats.PercentileNearestRank(buffer, 75)
		case "91%-ile":
			val, err = stats.PercentileNearestRank(buffer, 91)
		case "98%-ile":
			val, err = stats.PercentileNearestRank(buffer, 98)
		case "Minimum":
			val, err = min, minErr
		case "Maximum":
			val, err = max, maxErr
		case "Range":
			val = max - min
		case "Mode":
			modeVal, err = stats.Mode(buffer)
		case "Kurtosis":
			val, err = p.Kurtosis(buffer)
		case "Skewness":
			val, err = p.Skewness(buffer)
		case "Sum":
			val, err = stats.Sum(buffer)
		case "Trimean":
			val, err = stats.Trimean(buffer)
		case "Quartile_Range":
			val, err = stats.InterQuartileRange(buffer)
		default:
			st := fmt.Sprintf("Unknown statistic received %T:", stat)
			log.Errorf(st)
			err = errors.New(st)
		}

		if err != nil {
			log.Warnf("Error in %T", stat)
		}

		metric := plugin.Metric{
			Data:      val,
			Namespace: plugin.NewNamespace(namespace, stat),
			Timestamp: time,
			Unit:      unit,
			Tags:      tags,
		}

		if stat == "Mode" {
			metric.Data = modeVal
		}

		result = append(result, metric)

	}
	return result, err
}

//Calaculates the mean and standard deviation of a float64 array.
func (p *Plugin) MeanStdDev(buffer []float64) (float64, float64, error) {
	mean, err := stats.Mean(buffer)
	if err != nil {
		log.Warn(err)
		return math.NaN(), math.NaN(), err
	}

	stdev, err := stats.StandardDeviation(buffer)
	if err != nil {
		log.Warn(err)
		return mean, math.NaN(), err
	}

	return mean, stdev, nil
}

//Calculates the population skewness from buffer
func (p *Plugin) Skewness(buffer []float64) (float64, error) {
	if len(buffer) == 0 {
		log.Printf("Buffer does not contain any data.")
		return math.NaN(), errors.New("Buffer doesn't contain any data")
	}
	var skew float64
	mean, stdev, err := p.MeanStdDev(buffer)

	if err != nil {
		log.Fatal(err)
		return math.NaN(), err
	}

	for _, val := range buffer {
		skew += math.Pow((val-mean)/stdev, 3)
	}

	return float64(1 / float64(len(buffer)) * skew), nil

}

//Calculates the population kurtosis from buffer
func (p *Plugin) Kurtosis(buffer []float64) (float64, error) {
	if len(buffer) == 0 {
		log.Printf("Buffer does not contain any data.")
		return math.NaN(), errors.New("Buffer doesn't contain any data")
	}
	var kurt float64

	mean, stdev, err := p.MeanStdDev(buffer)
	if err != nil {
		log.Fatal(err)
		return math.NaN(), err
	}

	for _, val := range buffer {
		kurt += math.Pow((val-mean)/stdev, 4)
	}
	return float64(1 / float64(len(buffer)) * kurt), nil
}

// insertInToBuffer adds a new value into this' buffer object
func (p *Plugin) insertInToBuffer(val interface{}, ns string) {

	if p.bufferCurSize == 0 {
		var buff = make([]interface{}, p.bufferMaxSize)
		buff[0] = val
		p.buffer[ns] = buff
	} else {
		p.buffer[ns][p.bufferIndex] = val
	}
}

// updateCounters updates the meta informaiton (current size and index) of this' buffer object
func (p *Plugin) updateCounters() {
	if p.bufferCurSize < p.bufferMaxSize {
		p.bufferCurSize++
	}

	if p.bufferIndex == p.bufferMaxSize-1 {
		p.bufferIndex = 0
	} else {
		p.bufferIndex++
	}
}

// GetConfigPolicy returns the config policy
func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewIntRule([]string{""}, "SlidingWindowLength", false, plugin.SetDefaultInt(100))
	return *policy, nil

}

type byTimestamp []plugin.Metric

func (sa byTimestamp) Len() int {
	return len(sa)
}

func (sa byTimestamp) Less(i, j int) bool {
	return sa[i].Timestamp.Before(sa[j].Timestamp)

}
func (sa byTimestamp) Swap(i, j int) {
	sa[i], sa[j] = sa[j], sa[i]
}

func (p *Plugin) insertInToTimeBuffer(metric plugin.Metric, times []time.Time) []time.Time {
	times[p.bufferIndex] = metric.Timestamp
	return times
}

func (p *Plugin) getTimes(times []time.Time) (time.Time, time.Time) {
	if p.bufferCurSize == p.bufferMaxSize && p.bufferIndex != p.bufferMaxSize-1 {
		return times[p.bufferIndex+1], times[p.bufferIndex]
	}
	return times[0], times[p.bufferIndex]
}

// Process processes the data, inputs the data into this' buffer and calls the descriptive statistics method
func (p *Plugin) Process(metrics []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	f, err := os.OpenFile("/tmp/statisticErr.txt", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Warn("File reading error.")
	}

	log.SetOutput(f)

	bufsize, err := cfg.GetInt("SlidingWindowLength")
	p.bufferMaxSize = int(bufsize)
	if err != nil {
		return nil, err
	}

	var results []plugin.Metric
	times := make([]time.Time, p.bufferMaxSize)

	metricNamespace := make(map[string][]plugin.Metric)
	for _, metric := range metrics {
		//Populates the metricNamespace map so that the statistics are ran on metrics that share the same namespace.
		ns := metric.Namespace.String()
		if plugins, ok := metricNamespace[ns]; ok {
			plugins = append(plugins, metric)
			metricNamespace[ns] = plugins
		} else {
			metricNamespace[ns] = []plugin.Metric{metric}
		}
	}

	for k, v := range metricNamespace {
		var startTime time.Time
		var stopTime time.Time
		unit := v[0].Unit

		sort.Sort(byTimestamp(v))
		for _, metric := range v {
			times = p.insertInToTimeBuffer(metric, times)
			startTime, stopTime = p.getTimes(times)
			p.insertInToBuffer(metric.Data, metric.Namespace.String())
			p.updateCounters()

			if p.bufferCurSize < p.bufferMaxSize {
				stats, err := p.calculateStats(p.buffer[k][0:p.bufferCurSize], startTime, stopTime, k, unit)
				if err != nil {
					log.Warnf("Error occured in calculating Statistics: %s", err)
					return nil, err
				}
				results = append(results, stats...)
			} else if p.bufferCurSize == p.bufferMaxSize {
				log.Printf("Buffer: %v", p.buffer[k])
				stats, err := p.calculateStats(p.buffer[k], startTime, stopTime, k, unit)
				if err != nil {
					log.Warnf("Error occured in calculating Statistics: %s", err)
					return nil, err
				}
				results = append(results, stats...)
			}
		}
	}

	return results, nil
}
