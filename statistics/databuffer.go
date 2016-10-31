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

//dataBuffer contains 2 slices that point to buffer values:
type dataBuffer struct {
	data                      buffer //data is sorted by value
	dataByTimestamp           buffer //dataByTimestamp is sorted by timestamp
	index, slidingFactorIndex int    //sliding factor specifies how many data values to include over each sliding window
}
type buffer []*data

//data holds the timestamp and the value (actual data)
type data struct {
	ts    time.Time //timestamp
	value float64   //data
}

type Flag struct {
	Count, Sum, Mean, Median, Minimum, Maximum, Range, Variance,
	StandardDeviation, Mode, Kurtosis, Skewness, Trimean,
	FirstQuartile, ThirdQuartile, QuartileRange,
	SecondPercentile, NinthPercentile, TwentyFifthPercentile, SeventyFifthPercentile,
	NinetyFirstPercentile, NinetyFifthPercentile, NinetyEighthPercentile,
	NinetyNinthPercentile bool
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

//functions to sort data
func (a buffer) Len() int           { return len(a) }
func (a buffer) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a buffer) Less(i, j int) bool { return a[i].value < a[j].value }

//Inserts data in a sorted order
func (b *dataBuffer) Insert(value float64, ts time.Time) {
	if len(b.data) < cap(b.data) {
		element := &data{value: value, ts: ts}
		b.data = append(b.data, element)
		b.dataByTimestamp = append(b.dataByTimestamp, element)
	} else {
		b.dataByTimestamp[b.index].ts, b.dataByTimestamp[b.index].value = ts, value
		b.index = (b.index + 1) % cap(b.data)
	}
	sort.Sort(b.data)
}

//Set flags to true to be able to configure process functions
func SetFlags(stats []string) (flag Flag, err error) {
	for _, stat := range stats {
		switch stat {
		case count:
			flag.Count = true
		case mean:
			flag.Mean = true
		case median:
			flag.Median = true
		case standarddeviation:
			flag.StandardDeviation = true
		case variance:
			flag.Variance = true
		case ninetyfifthpercentile:
			flag.NinetyFifthPercentile = true
		case ninetyninthpercentile:
			flag.NinetyNinthPercentile = true
		case secondpercentile:
			flag.SecondPercentile = true
		case ninthpercentile:
			flag.NinthPercentile = true
		case twentyfifthpercentile:
			flag.TwentyFifthPercentile = true
		case seventyfifthpercentile:
			flag.SeventyFifthPercentile = true
		case ninetyfirstpercentile:
			flag.NinetyFirstPercentile = true
		case ninetyeighthpercentile:
			flag.NinetyEighthPercentile = true
		case minimum:
			flag.Minimum = true
		case maximum:
			flag.Maximum = true
		case rangeval:
			flag.Range = true
		case mode:
			flag.Mode = true
		case kurtosis:
			flag.Kurtosis = true
		case skewness:
			flag.Skewness = true
		case sum:
			flag.Sum = true
		case trimean:
			flag.Trimean = true
		case quartilerange:
			flag.QuartileRange = true
		case firstquartile:
			flag.FirstQuartile = true
		case thirdquartile:
			flag.ThirdQuartile = true
		default:
			err = fmt.Errorf("Unknown statistic received %T:", stat)
			return
		}
	}
	return
}

// Creates a metric for each statistic
func createMetric(data interface{}, tags map[string]string) plugin.Metric {
	return plugin.Metric{
		Timestamp: time.Now(),
		Tags:      tags,
		Data:      data,
	}
}

// Get tags which are start time and stop time
func (b *dataBuffer) GetTags() map[string]string {
	new := (b.index + len(b.data) - 1) % len(b.data)
	old := b.index
	oldTs := b.dataByTimestamp[old].ts
	newTs := b.dataByTimestamp[new].ts
	return map[string]string{"startTime": oldTs.String(), "stopTime": newTs.String()}
}

// Checks if flag is true and calls the appropriate process functions
func (d *dataBuffer) GetStats(stats []string, ns []string) ([]plugin.Metric, error) {
	var results []plugin.Metric
	var metric plugin.Metric
	nsPrefix := []string{"intel", "statistics"}
	tags := d.GetTags()

	var result struct {
		sum, mean, median, minimum, maximum, Range, variance,
		standarddeviation, kurtosis, skewness, trimean,
		firstquartile, thirdquartile, quartilerange,
		secondpercentile, ninthpercentile, twentyfifthpercentile, seventyfifthpercentile,
		ninetyfirstpercentile, ninetyfifthpercentile, ninetyeighthpercentile,
		ninetyninthpercentile float64
		count int
	}

	if len(d.data) == 0 {
		return nil, nil
	}
	flag, err := SetFlags(stats)
	if err != nil {
		return nil, err
	}

	if flag.Count {
		result.count = d.Count()
		metric = createMetric(result.count, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(count)
		results = append(results, metric)
	}

	if flag.Sum {
		result.sum = d.Sum()
		metric = createMetric(result.sum, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(sum)
		results = append(results, metric)
	}

	if flag.Minimum || flag.Range {
		result.minimum = d.Minimum()
		if flag.Minimum {
			metric = createMetric(result.minimum, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(minimum)
			results = append(results, metric)
		}
	}

	if flag.Maximum || flag.Range {
		result.maximum = d.Maximum()
		if flag.Maximum {
			metric = createMetric(result.maximum, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(maximum)
			results = append(results, metric)
		}
	}

	if flag.Range {
		result.Range = d.Range(result.minimum, result.maximum)
		metric = createMetric(result.Range, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(rangeval)
		results = append(results, metric)
	}

	if flag.Mean || flag.Kurtosis || flag.Skewness || flag.Variance {
		result.mean = d.Mean(result.sum, result.count)
		if flag.Mean {
			metric = createMetric(result.mean, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(mean)
			results = append(results, metric)
		}
	}

	if flag.Median || flag.Trimean {
		result.median = d.Median()
		if flag.Median {
			metric = createMetric(result.median, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(median)
			results = append(results, metric)
		}
	}
	if flag.Mode {
		modes := d.Mode()
		for _, val := range modes {
			metric = createMetric(val, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(mode).AddDynamicElement("highestfreq", "Gives the highest number of occurences of a data value")
			results = append(results, metric)
		}
	}

	if flag.FirstQuartile || flag.QuartileRange || flag.Trimean {
		result.firstquartile = d.FirstQuartile()
		if flag.FirstQuartile {
			metric = createMetric(result.firstquartile, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(firstquartile)
			results = append(results, metric)
		}
	}
	if flag.ThirdQuartile || flag.QuartileRange || flag.Trimean {
		result.thirdquartile = d.ThirdQuartile()
		if flag.ThirdQuartile {
			metric = createMetric(result.thirdquartile, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(thirdquartile)
			results = append(results, metric)
		}
	}

	if flag.QuartileRange {
		result.quartilerange = d.Range(result.firstquartile, result.thirdquartile)
		metric = createMetric(result.quartilerange, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(quartilerange)
		results = append(results, metric)
	}

	if flag.Variance || flag.StandardDeviation || flag.Skewness || flag.Kurtosis {
		result.variance = d.Variance(result.mean)
		if flag.Variance {
			metric = createMetric(result.variance, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(variance)
			results = append(results, metric)
		}
	}

	if flag.StandardDeviation {
		result.standarddeviation = d.StandardDeviation(result.variance)
		metric = createMetric(result.standarddeviation, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(standarddeviation)
		results = append(results, metric)
	}

	if flag.SecondPercentile {
		result.secondpercentile, err = d.PercentileNearestRank(2)
		if err != nil {
			return nil, err
		}

		metric = createMetric(result.secondpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(secondpercentile)
		results = append(results, metric)
	}

	if flag.NinthPercentile {
		result.ninthpercentile, err = d.PercentileNearestRank(9)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.ninthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(ninthpercentile)
		results = append(results, metric)
	}

	if flag.TwentyFifthPercentile {
		result.twentyfifthpercentile, err = d.PercentileNearestRank(25)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.twentyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(twentyfifthpercentile)
		results = append(results, metric)
	}

	if flag.SeventyFifthPercentile {
		result.seventyfifthpercentile, err = d.PercentileNearestRank(75)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.seventyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(seventyfifthpercentile)
		results = append(results, metric)
	}

	if flag.NinetyFirstPercentile {
		result.ninetyfirstpercentile, err = d.PercentileNearestRank(91)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.ninetyfirstpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(ninetyfirstpercentile)
		results = append(results, metric)
	}

	if flag.NinetyEighthPercentile {
		result.ninetyeighthpercentile, err = d.PercentileNearestRank(98)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.ninetyeighthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(ninetyeighthpercentile)
		results = append(results, metric)
	}

	if flag.NinetyFifthPercentile {
		result.ninetyfifthpercentile, err = d.PercentileNearestRank(95)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.ninetyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(ninetyfifthpercentile)
		results = append(results, metric)
	}

	if flag.NinetyNinthPercentile {
		result.ninetyninthpercentile, err = d.PercentileNearestRank(99)
		if err != nil {
			return nil, err
		}
		metric = createMetric(result.ninetyninthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(ninetyninthpercentile)
		results = append(results, metric)
	}

	if flag.Skewness {
		result.skewness = d.Skewness(result.mean, result.standarddeviation)
		if result.skewness != math.NaN() {
			metric = createMetric(result.skewness, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(skewness)
			results = append(results, metric)
		}
	}

	if flag.Kurtosis {
		result.kurtosis = d.Kurtosis(result.mean, result.standarddeviation)
		if result.kurtosis != math.NaN() {
			metric = createMetric(result.kurtosis, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(kurtosis)
			results = append(results, metric)
		}
	}

	if flag.Trimean {
		result.trimean = d.Trimean(result.firstquartile, result.median, result.thirdquartile)
		if result.trimean != math.NaN() {
			metric = createMetric(result.trimean, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement(trimean)
			results = append(results, metric)
		}
	}
	return results, err
}
