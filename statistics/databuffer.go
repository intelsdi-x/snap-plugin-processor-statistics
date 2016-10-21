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
	data            buffer
	dataByTimestamp buffer
	index           int
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
			flag.Variance, flag.StandardDeviation = true, true
		case "variance":
			flag.Mean = true
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
			//flag.Maximum = true
			//flag.Minimum = true
			flag.Range = true
		case "mode":
			flag.Mode = true
		case "kurtosis":
			flag.Mean = true
			flag.StandardDeviation = true
			flag.Kurtosis = true
		case "skewness":
			flag.Mean = true
			flag.StandardDeviation = true
			flag.Skewness = true
		case "sum":
			flag.Sum = true
		case "trimean":
			flag.FirstQuartile = true
			flag.Median = true
			flag.ThirdQuartile = true
			flag.Trimean = true
		case "quartile_range":
			//flag.FirstQuartile = true
			//flag.ThirdQuartile = true
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
	bufferSize := len(d.data)
	var (
		count, sum, mean, median, minimum, maximum, Range, variance,
		standarddeviation, mode, kurtosis, skewness, trimean,
		firstquartile, thirdquartile, quartilerange,
		secondpercentile, ninthpercentile, twentyfifthpercentile, seventyfifthpercentile,
		ninetyfirstpercentile, ninetyfifthpercentile, ninetyeighthpercentile,
		ninetyninthpercentile float64
	)
	flag, err := SetFlags(stats)
	if err != nil {
		return nil, err
	}

	if flag.Sum {
		sum = d.Sum()
		metric = createMetric(sum, tags)
		metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("sum")
		results = append(results, metric)
	}

	if flag.Range || flag.Minimum {
		minimum = d.Minimum()
		if flag.Minimum {
			metric = createMetric(sum, tags)
			metric.Namespace = plugin.NewNamespace(nsPrefix...).AddStaticElements(ns...).AddStaticElement("minimum")
			results = append(results, metric)
		}
	}

	if flag.Range || flag.Maximum {
		maximum = d.Maximum()
		if flag.Maximum {
			ns := plugin.NewNamespace(nsPrefix).AddStaticElements(ns).AddStaticElement("maximum")
			results = append(results, createMetric(maximum))
		}
	}
	if flag.Range {
		Range = d.Range(Minimum, Maximum)
		// create metric range
	}

	if flag.Mean {
		Mean = d.Mean(Sum, Count)
		// create metric mean
	}

	if flag.Median {
		Median = d.Median()
		//create metric median
	}

	if flag.QuartileRange || flag.FirstQuartile {
		FirstQuartile = d.FirstQuartile()
		if flag.FirstQuartile {
			// create metric first quartile
		}
	}

	if flag.QuartileRange || flag.ThirdQuartile {
		ThirdQuartile = d.ThirdQuartile()
		if flag.ThirdQuartile {
			// create metric third quartile
		}
	}
	if flag.QuartileRange {
		QuartileRange = d.Range(FirstQuartile, ThirdQuartile)
		// create metric range
	}

	if flag.Variance || flag.Mean {
		Mean = d.Mean(Sum, Count)
		if flag.Mean {
			// create metric mean
		}
	}

	if flag.Variance {
		Variance = d.Variance(Mean)
		// create metric variance
	}

	if flag.StandardDeviation || flag.Variance {
		Variance = d.Variance(Mean)
		if flag.Variance {
			// create metric variance
		}
	}

	if flag.SecondPercentile {
		SecondPercentile = d.PercentileNearestRank(2)
		// create metric second percentile
	}

	if flag.NinthPercentile {
		NinthPercentile = d.PercentileNearestRank(9)
		// create metric ninth percentile
	}

	if flag.TwentyFifthPercentile {
		TwentyFifthPercentile = d.PercentileNearestRank(25)
		// create metric twenty fifth percentile
	}

	if flag.SeventyFifthPercentile {
		SeventyFifthPercentile = d.PercentileNearestRank(75)
		// create metric seventy fifth percentile
	}

	if flag.NinetyFirstPercentile {
		NinetyFirstPercentile = d.PercentileNearestRank(91)
		// create metric ninety first percentile
	}

	if flag.NinetyEighthPercentile {
		NinetyEighthPercentile = d.PercentileNearestRank(98)
		// create metric ninety eighth percentile
	}

	if flag.NinetyFifthPercentile {
		NinetyFifthPercentile = d.PercentileNearestRank(95)
		// create metric ninety first percentile
	}

	if flag.NinetyNinthPercentile {
		NinetyNinthPercentile = d.PercentileNearestRank(99)
		// create metric ninety ninth percentile
	}

	if flag.Skewness {
		Skewness = d.Skewness(Mean, StandardDeviation)
		// create metric skewness percentile
	}

	if flag.Kurtosis {
		Kurtosis = d.Kurtosis(Mean, StandardDeviation)
		// create metric kurtosis percentile
	}

	// metric := plugin.Metric{
	// 	Data:      val,
	// 	Namespace: plugin.NewNamespace(namespace, stat).AddDynamicElement("Window count", "This is the Nth window"),
	// 	Timestamp: time,
	// 	Unit:      unit,
	// 	Tags:      tags,
	// }

	// results = append(results, metric)

	//return
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
	//return the last item
	// if percent == 100.0 {
	// 	return d.data[l-1].value
	// }

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
