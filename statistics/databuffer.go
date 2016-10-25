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

//Contains a slice of data
type dataBuffer struct {
	data                      buffer
	dataByTimestamp           buffer
	index, slidingFactorIndex int
}
type buffer []*data

//data holds the timestamp and the value (actual data)
type data struct {
	ts    time.Time
	value float64
}

// Quartiles holds the three quartile points
type Quartiles struct {
	Q1 float64
	Q2 float64
	Q3 float64
}

type Flag struct {
	Count, Sum, Mean, Median, Minimum, Maximum, Range, Variance,
	StandardDeviation, Mode, Kurtosis, Skewness, Trimean,
	FirstQuartile, ThirdQuartile, QuartileRange,
	SecondPercentile, NinthPercentile, TwentyFifthPercentile, SeventyFifthPercentile,
	NinetyFirstPercentile, NinetyFifthPercentile, NinetyEighthPercentile,
	NinetyNinthPercentile bool
}

func (a buffer) Len() int           { return len(a) }
func (a buffer) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a buffer) Less(i, j int) bool { return a[i].value < a[j].value }

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

func (b *dataBuffer) timeDiff() time.Duration {
	new := (b.index + len(b.data) - 1) % len(b.data)
	old := b.index
	return b.dataByTimestamp[new].ts.Sub(b.dataByTimestamp[old].ts)
}

func SetFlags(stats []string) (flag Flag, err error) {
	for _, stat := range stats {
		switch stat {
		case "count":
			flag.Count = true
		case "mean":
			flag.Mean = true
		case "median":
			flag.Median = true
		case "standard_deviation":
			flag.StandardDeviation = true
		case "variance":
			flag.Variance = true
		case "95%_ile":
			flag.NinetyFifthPercentile = true
		case "99%_ile":
			flag.NinetyNinthPercentile = true
		case "2%_ile":
			flag.SecondPercentile = true
		case "9%_ile":
			flag.NinthPercentile = true
		case "25%_ile":
			flag.TwentyFifthPercentile = true
		case "75%_ile":
			flag.SeventyFifthPercentile = true
		case "91%_ile":
			flag.NinetyFirstPercentile = true
		case "98%_ile":
			flag.NinetyEighthPercentile = true
		case "minimum":
			flag.Minimum = true
		case "maximum":
			flag.Maximum = true
		case "range":
			flag.Range = true
		case "mode":
			flag.Mode = true
		case "kurtosis":
			flag.Kurtosis = true
		case "skewness":
			flag.Skewness = true
		case "sum":
			flag.Sum = true
		case "trimean":
			flag.Trimean = true
		case "quartile_range":
			flag.QuartileRange = true
		case "first_quartile":
			flag.FirstQuartile = true
		case "third_quartile":
			flag.ThirdQuartile = true
		default:
			err = fmt.Errorf("Unknown statistic received %T:", stat)
			return
		}
	}
	return
}

func createMetric(data float64, tags map[string]string) plugin.Metric {
	return plugin.Metric{
		Timestamp: time.Now(),
		Tags:      tags,
		Data:      data,
	}
}

func (b *dataBuffer) GetTags() map[string]string {
	new := (b.index + len(b.data) - 1) % len(b.data)
	old := b.index
	oldTs := b.dataByTimestamp[old].ts
	newTs := b.dataByTimestamp[new].ts
	return map[string]string{"startTime": oldTs.String(), "stopTime": newTs.String()}
}

func (d *dataBuffer) GetStats(stats []string, ns []string) ([]plugin.Metric, error) {
	var results []plugin.Metric
	var metric plugin.Metric
	nsPrefix := []string{"intel", "statistics"}
	tags := d.GetTags()
	var (
		count, sum, mean, median, minimum, maximum, Range, variance,
		standarddeviation, kurtosis, skewness, trimean,
		firstquartile, thirdquartile, quartilerange,
		secondpercentile, ninthpercentile, twentyfifthpercentile, seventyfifthpercentile,
		ninetyfirstpercentile, ninetyfifthpercentile, ninetyeighthpercentile,
		ninetyninthpercentile float64
	)
	var mode []float64

	flag, err := SetFlags(stats)
	if err != nil {
		return nil, err
	}

	if flag.Mode {
		mode = d.Mode()
		for _, val := range mode {
			metric = createMetric(val, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("mode").AddDynamicElement("highestfreq", "Gives the highest number of occurences of a data value")
			results = append(results, metric)
		}
	}

	if flag.Sum {
		sum = d.Sum()
		metric = createMetric(sum, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("sum")
		results = append(results, metric)
	}

	if flag.Minimum || flag.Range {
		minimum = d.Minimum()
		if flag.Minimum {
			metric = createMetric(minimum, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("minimum")
			results = append(results, metric)
		}
	}

	if flag.Maximum || flag.Range {
		maximum = d.Maximum()
		if flag.Maximum {
			metric = createMetric(maximum, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("maximum")
			results = append(results, metric)
		}
	}
	if flag.Range {
		Range = d.Range(minimum, maximum)
		metric = createMetric(Range, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("range")
		results = append(results, metric)
	}

	if flag.Mean || flag.Kurtosis || flag.Skewness {
		mean = d.Mean(sum, count)
		if flag.Mean {
			metric = createMetric(mean, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("mean")
			results = append(results, metric)
		}
	}

	if flag.Median || flag.Trimean {
		median = d.Median()
		if flag.Median {
			metric = createMetric(median, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("median")
			results = append(results, metric)
		}
	}

	if flag.FirstQuartile || flag.QuartileRange || flag.Trimean {
		firstquartile = d.FirstQuartile()
		if flag.FirstQuartile {
			metric = createMetric(firstquartile, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("firstquartile")
			results = append(results, metric)
		}
	}

	if flag.ThirdQuartile || flag.QuartileRange || flag.Trimean {
		thirdquartile = d.ThirdQuartile()
		if flag.ThirdQuartile {
			metric = createMetric(thirdquartile, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("thirdquartile")
			results = append(results, metric)
		}
	}
	if flag.QuartileRange {
		quartilerange = d.Range(firstquartile, thirdquartile)
		metric = createMetric(quartilerange, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("quartilerange")
		results = append(results, metric)
	}

	if flag.Mean || flag.Variance {
		mean = d.Mean(sum, count)
		if flag.Mean {
			metric = createMetric(mean, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("mean")
			results = append(results, metric)
		}
	}

	if flag.Variance || flag.StandardDeviation || flag.Skewness || flag.Kurtosis {
		variance = d.Variance(mean)
		if flag.Variance {
			metric = createMetric(variance, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("variance")
			results = append(results, metric)
		}
	}

	if flag.SecondPercentile {
		secondpercentile = d.PercentileNearestRank(2)
		metric = createMetric(secondpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("secondpercentile")
		results = append(results, metric)
	}

	if flag.NinthPercentile {
		ninthpercentile = d.PercentileNearestRank(9)
		metric = createMetric(ninthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("ninthpercentile")
		results = append(results, metric)
	}

	if flag.TwentyFifthPercentile {
		twentyfifthpercentile = d.PercentileNearestRank(25)
		metric = createMetric(twentyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("twentyfifthpercentile")
		results = append(results, metric)
	}

	if flag.SeventyFifthPercentile {
		seventyfifthpercentile = d.PercentileNearestRank(75)
		metric = createMetric(seventyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("seventyfifthpercentile")
		results = append(results, metric)
	}

	if flag.NinetyFirstPercentile {
		ninetyfirstpercentile = d.PercentileNearestRank(91)
		metric = createMetric(ninetyfirstpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("ninetyfirstpercentile")
		results = append(results, metric)
	}

	if flag.NinetyEighthPercentile {
		ninetyeighthpercentile = d.PercentileNearestRank(98)
		metric = createMetric(ninetyeighthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("ninetyeighthpercentile")
		results = append(results, metric)
	}

	if flag.NinetyFifthPercentile {
		ninetyfifthpercentile = d.PercentileNearestRank(95)
		metric = createMetric(ninetyfifthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("ninetyfifthpercentile")
		results = append(results, metric)
	}

	if flag.NinetyNinthPercentile {
		ninetyninthpercentile = d.PercentileNearestRank(99)
		metric = createMetric(ninetyninthpercentile, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("ninetyninthpercentile")
		results = append(results, metric)
	}

	if flag.Skewness {
		skewness = d.Skewness(mean, standarddeviation)
		metric = createMetric(skewness, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("skewness")
		results = append(results, metric)
	}

	if flag.Kurtosis {
		kurtosis = d.Kurtosis(mean, standarddeviation)
		metric = createMetric(kurtosis, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("kurtosis")
		results = append(results, metric)
	}

	if flag.Trimean {
		trimean = d.Trimean(firstquartile, median, thirdquartile)
		metric = createMetric(trimean, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("trimean")
		results = append(results, metric)
	}
	return results, err
}

func (d *dataBuffer) Sum() (sum float64) {
	for _, val := range d.data {
		sum += val.value
	}
	return
}

func (d *dataBuffer) Mean(sum, count float64) float64 {
	return sum / count
}

func (d *dataBuffer) Minimum() (min float64) {
	for _, val := range d.data {
		if val.value < min {
			min = val.value
		}
	}
	return
}

func (d *dataBuffer) Maximum() (max float64) {
	for _, val := range d.data {
		if val.value > max {
			max = val.value
		}
	}
	return
}

func (d *dataBuffer) Range(min, max float64) float64 {
	return (max - min)
}
func (d *dataBuffer) Variance(mean float64) (variance float64) {
	var total float64
	for _, val := range d.data {
		total += math.Pow(val.value-mean, 2)
	}
	variance = total / float64(len(d.data)-1)
	return
}

func (d *dataBuffer) StandardDeviation(variance float64) float64 {
	return math.Sqrt(variance)
}

func (d *dataBuffer) Median() (median float64) {
	l := len(d.data)
	if l == 0 {
		return math.NaN()
	} else if l%2 == 0 {
		median = (d.data[l/2-1].value + d.data[l/2].value) / 2
	} else {
		median = d.data[l/2].value
	}
	return
}

func (d *dataBuffer) PercentileNearestRank(percent float64) float64 {
	l := len(d.data)
	//return error less than 0 or greater than 100 percentages
	if percent < 0 || percent > 100 {
		return math.NaN()
	}

	//find the ordinal ranking
	or := int(math.Ceil(float64(l) * percent / 100.0))

	//return the item that is in the place of the ordinal rank
	if or == 0 {
		return d.data[0].value
	}
	return d.data[or-1].value

}

func (d *dataBuffer) Mode() (modes []float64) {
	frequencies := make(map[float64]int, len(d.data))
	highestFrequency := 0
	for _, x := range d.data {
		frequencies[x.value]++
		if frequencies[x.value] > highestFrequency {
			highestFrequency = frequencies[x.value]
		}
	}
	for x, frequency := range frequencies {
		if frequency == highestFrequency {
			modes = append(modes, x)
		}
	}
	if highestFrequency == 1 || len(modes) == len(d.data) {
		modes = modes[:0]
	}
	return
}

// Quartile returns the three quartile points from a slice of data
func (d *dataBuffer) FirstQuartile() (quartile float64) {
	l := len(d.data)
	if l == 0 {
		return math.NaN()
	}

	//find the cutoff places depending on if the input slice length is even or odd
	if l%2 == 0 {
		l = l / 2
	} else {
		l = (l - 1) / 2
	}
	if l == 0 {
		return math.NaN()
	} else if l%2 == 0 {
		quartile = (d.data[l/2-1].value + d.data[l/2].value) / 2
	} else {
		quartile = d.data[l/2].value
	}

	return
}

// Quartile returns the three quartile points from a slice of data
func (d *dataBuffer) ThirdQuartile() (quartile float64) {
	l := len(d.data)
	if l == 0 {
		return math.NaN()
	}

	//find the cutoff places depending on if the input slice length is even or odd
	var c1 int
	if l%2 == 0 {
		l = l / 2
	} else {
		c1 = (l - 1) / 2
		l = c1 + 1
	}

	if l == 0 {
		return math.NaN()
	} else if l%2 == 0 {
		quartile = (d.data[l/2-1].value + d.data[l/2].value) / 2
	} else {
		quartile = d.data[l/2].value
	}
	quartile = (d.data[c1/2-1].value + d.data[c1/2].value) / 2

	return
}

// InterQuartileRange finds the range between Q1 and Q3
func (d *dataBuffer) QuartileRange(firstquartile, thirdquartile float64) float64 {
	if len(d.data) == 0 {
		return math.NaN()
	}
	return thirdquartile - firstquartile
}

// Trimean finds the average of the median and the midhinge
func (d *dataBuffer) Trimean(firstquartile, median, thirdquartile float64) float64 {
	if len(d.data) == 0 {
		return math.NaN()
	}

	return (firstquartile + (median * 2) + thirdquartile) / 4

}

//Calculates the population skewness from buffer
func (d *dataBuffer) Skewness(mean, stdev float64) (skew float64) {
	l := len(d.data)
	if l == 0 {
		return math.NaN()
	}

	for _, val := range d.data {
		skew += math.Pow((val.value-mean)/stdev, 3)
	}

	return 1.0 / float64(l) * skew

}

//Calculates the population kurtosis from buffer
func (d *dataBuffer) Kurtosis(mean, stdev float64) (kurt float64) {
	l := len(d.data)
	if l == 0 {
		return math.NaN()
	}

	for _, val := range d.data {
		kurt += math.Pow((val.value-mean)/stdev, 4)
	}
	return 1.0 / float64(l) * kurt
}
