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
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type dataBuffer struct {
	data                         []data
	slidingFactorIndex, old, new int //sliding factor specifies how many data values to include over each sliding window
}

// data holds the timestamp and the value (actual data)
type data struct {
	ts    time.Time
	value float64
}

const (
	count                  = "count"
	mean                   = "mean"
	sum                    = "sum"
	median                 = "median"
	minimum                = "minimum"
	maximum                = "maximum"
	rangeval               = "rangeval"
	variance               = "variance"
	standarddeviation      = "standarddeviation"
	mode                   = "mode"
	kurtosis               = "kurtosis"
	skewness               = "skewness"
	trimean                = "trimean"
	firstquartile          = "firstquartile"
	thirdquartile          = "thirdquartile"
	quartilerange          = "quartilerange"
	secondpercentile       = "secondpercentile"
	ninthpercentile        = "ninthpercentile"
	twentyfifthpercentile  = "twentyfifthpercentile"
	seventyfifthpercentile = "seventyfifthpercentile"
	ninetyfirstpercentile  = "ninetyfirstpercentile"
	ninetyeighthpercentile = "ninetyeighthpercentile"
	ninetyninthpercentile  = "ninetyninthpercentile"
	ninetyfifthpercentile  = "ninetyfifthpercentile"
)

var (
	statList = []string{count, mean, sum, median, minimum, maximum, rangeval, variance, standarddeviation, mode, kurtosis, skewness, trimean, firstquartile, thirdquartile, quartilerange,
		secondpercentile, ninthpercentile, twentyfifthpercentile, seventyfifthpercentile, ninetyfirstpercentile, ninetyeighthpercentile, ninetyfifthpercentile, ninetyninthpercentile}
)

func (b *dataBuffer) Insert(value float64, ts time.Time) {
	// sort by timestamp before inserting
	sort.Sort(byTimestamp(b.data))
	if len(b.data) < cap(b.data) {
		b.data = append(b.data, data{value: value, ts: ts})
	} else {
		// replace older value
		b.data[0].ts, b.data[0].value = ts, value
	}
}

func (d *dataBuffer) GetStats(stats []string, ns []string) ([]plugin.Metric, error) {
	if len(d.data) == 0 {
		return nil, nil
	}
	var results []plugin.Metric

	// Namespace prefix and tags are common for every stats
	nsPrefix := append([]string{"intel", "statistics"}, ns...)
	tags := d.GetTags()

	// statistics are stored in a map
	statMap := make(result)

	// sort by value is required before calculating statistics
	sort.Sort(byValue(d.data))

	// generate config option map
	opts, err := d.SetConfigOption(stats)
	if err != nil {
		return nil, err
	}

	// Calcul statistics
	for stat, opt := range opts {

		if _, ok := statMap[stat]; !ok {
			opt(statMap)
		}
		newStat := statMap[stat]

		// create the metric from the statistic we just calculated
		err = createMetrics(&results, newStat, tags, append(nsPrefix, stat))
		if err != nil {
			return nil, err
		}
	}

	return results, err
}

type byValue []data

//functions to sort by value
func (a byValue) Len() int           { return len(a) }
func (a byValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byValue) Less(i, j int) bool { return a[i].value < a[j].value }

type byTimestamp []data

//functions to sort by timestamp
func (a byTimestamp) Len() int           { return len(a) }
func (a byTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTimestamp) Less(i, j int) bool { return a[i].ts.After(a[j].ts) }

// Creates a metric for each statistic
func createMetrics(result *[]plugin.Metric, data interface{}, tags map[string]string, ns []string) error {
	switch data.(type) {
	case float64:
		if math.IsNaN(data.(float64)) {
			return nil
		}
	case int:
		// nothing to change
	case []float64:
		for _, val := range data.([]float64) {
			*result = append(*result, createMetric(val, tags,
				plugin.NewNamespace(ns...).AddDynamicElement("highestfreq", "Gives the highest number of occurences of a data value")))
		}
		return nil
	default:
		return fmt.Errorf("invalid type for a statistic")
	}

	namespace := plugin.NewNamespace(ns...)
	*result = append(*result, createMetric(data, tags, namespace))
	return nil
}

func createMetric(data interface{}, tags map[string]string, namespace plugin.Namespace) plugin.Metric {
	return plugin.Metric{
		Timestamp: time.Now(),
		Tags:      tags,
		Data:      data,
		Namespace: namespace,
	}
}

// Get tags which are start time and stop time
func (b *dataBuffer) GetTags() map[string]string {
	oldTs := b.data[0].ts
	newTs := b.data[len(b.data)-1].ts
	return map[string]string{"startTime": oldTs.String(), "stopTime": newTs.String()}
}
