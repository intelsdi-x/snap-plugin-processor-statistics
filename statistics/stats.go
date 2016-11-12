package statistics

import (
	"fmt"
	"math"
)

// Returns count of the buffer
func (d *dataBuffer) Count() int {
	return len(d.data)
}

// Returns sum of all the data values in the buffer
func (d *dataBuffer) Sum() (sum float64) {
	for _, val := range d.data {
		sum += val.value
	}
	return
}

// Returns the mean of the buffer
func (d *dataBuffer) Mean(sum float64, count int) float64 {
	return sum / float64(count)
}

// Returns the minimum value in the buffer
func (d *dataBuffer) Minimum() float64 {
	return d.data[0].value
}

// Returns the maximum value in the buffer
func (d *dataBuffer) Maximum() float64 {

	return d.data[len(d.data)-1].value
}

// Returns the range of the data buffer
func (d *dataBuffer) Range(min, max float64) float64 {
	return (max - min)
}

// Calculates and returns variance from mean
func (d *dataBuffer) Variance(mean float64) (variance float64) {
	var total float64
	l := len(d.data)
	if l == 1 {
		return 0
	}

	for _, val := range d.data {
		total += math.Pow(val.value-mean, 2)
	}
	variance = total / float64(len(d.data))
	return
}

// Calculates and returns standard deviation from variance
func (d *dataBuffer) StandardDeviation(variance float64) float64 {
	return math.Sqrt(variance)
}

// Returns median of the data buffer
func (d *dataBuffer) Median() (median float64) {
	l := len(d.data)
	if l%2 == 0 {
		median = (d.data[l/2-1].value + d.data[l/2].value) / 2
	} else {
		median = d.data[l/2].value
	}
	return
}

// Calculates the percentile based on the percent that is being passed as input
func (d *dataBuffer) PercentileNearestRank(percent float64) (float64, error) {
	l := len(d.data)
	//return error less than 0 or greater than 100 percentages
	if percent < 0 || percent > 100 {
		return 0, fmt.Errorf("not a valid percent")
	}

	//find the ordinal ranking
	or := int(math.Ceil(float64(l) * percent / 100.0))

	//return the item that is in the place of the ordinal rank
	if or == 0 {
		return d.data[0].value, nil
	}
	return d.data[or-1].value, nil
}

// Returns mode of the data buffer
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

// First quartile returns the first quartile point which is the middle number between the smallest number and the median of the data set
func (d *dataBuffer) FirstQuartile() (quartile float64) {
	l := len(d.data)

	//find the cutoff places depending on if the input slice length is even or odd
	if l%2 == 0 {
		l = l / 2
	} else {
		l = (l - 1) / 2
	}
	if l == 0 {
		return d.data[0].value
	} else if l%2 == 0 {
		quartile = (d.data[l/2-1].value + d.data[l/2].value) / 2
	} else {
		quartile = d.data[l/2].value
	}

	return
}

/* Third quartile returns the third quartile point from a slice of data,
which is the middle value between the median and the highest value of the data set */
func (d *dataBuffer) ThirdQuartile() (quartile float64) {
	l := len(d.data)
	if l == 1 {
		return d.data[0].value
	} else if l == 2 {
		return d.data[0].value*0.25 + d.data[1].value*0.75
	} else {
		c1 := l / 2
		c2 := l - 1
		l = (c2 - c1) + 1
		if l%2 == 0 {
			quartile = (d.data[c1+(l/2)-1].value + d.data[c1+(l/2)].value) / 2
		} else {
			quartile = d.data[c1+(l/2)].value
		}
	}

	return
}

// InterQuartileRange finds the range between first quartile and third quartile
func (d *dataBuffer) QuartileRange(firstquartile, thirdquartile float64) float64 {
	return thirdquartile - firstquartile
}

// Trimean finds the average of the median and the midhinge
func (d *dataBuffer) Trimean(firstquartile, median, thirdquartile float64) float64 {
	return (firstquartile + (median * 2) + thirdquartile) / 4

}

//Calculates the population skewness from the data buffer
func (d *dataBuffer) Skewness(mean, stdev float64) (skew float64) {
	l := len(d.data)

	for _, val := range d.data {
		skew += math.Pow((val.value-mean)/stdev, 3)
	}

	return 1.0 / float64(l) * skew
}

//Calculates the population kurtosis from the data buffer
func (d *dataBuffer) Kurtosis(mean, stdev float64) (kurt float64) {
	l := len(d.data)

	for _, val := range d.data {
		kurt += math.Pow((val.value-mean)/stdev, 4)
	}
	return 1.0 / float64(l) * kurt
}
