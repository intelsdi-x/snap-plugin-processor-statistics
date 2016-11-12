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

import "fmt"

type result map[string]interface{}
type statOpt func(result)

//Set flags to true to be able to configure process functions
func (d *dataBuffer) SetConfigOption(stats []string) (map[string]statOpt, error) {
	statOpts := make(map[string]statOpt)
	for _, stat := range stats {
		switch stat {
		case count:
			statOpts[count] = d.countOpt
		case mean:
			statOpts[mean] = d.meanOpt
		case median:
			statOpts[median] = d.medianOpt
		case standarddeviation:
			statOpts[standarddeviation] = d.standardDeviationOpt
		case variance:
			statOpts[variance] = d.varianceOpt
		case ninetyfifthpercentile:
			statOpts[ninetyfifthpercentile] = d.ninetyFifthPercentileOpt
		case ninetyninthpercentile:
			statOpts[ninetyninthpercentile] = d.ninetyNinthPercentileOpt
		case secondpercentile:
			statOpts[secondpercentile] = d.secondPercentileOpt
		case ninthpercentile:
			statOpts[ninthpercentile] = d.ninthPercentileOpt
		case twentyfifthpercentile:
			statOpts[twentyfifthpercentile] = d.twentyPercentileOpt
		case seventyfifthpercentile:
			statOpts[seventyfifthpercentile] = d.SeventyFifthPercentileOpt
		case ninetyfirstpercentile:
			statOpts[ninetyfirstpercentile] = d.ninetyFirstPercentileOpt
		case ninetyeighthpercentile:
			statOpts[ninetyeighthpercentile] = d.ninetyEightPercentileOpt
		case minimum:
			statOpts[minimum] = d.minimumOpt
		case maximum:
			statOpts[maximum] = d.maximumOpt
		case rangeval:
			statOpts[rangeval] = d.rangeOpt
		case mode:
			statOpts[mode] = d.modesOpt
		case kurtosis:
			statOpts[kurtosis] = d.kurtosisOpt
		case skewness:
			statOpts[skewness] = d.skewnessOpt
		case sum:
			statOpts[sum] = d.sumOpt
		case trimean:
			statOpts[trimean] = d.trimeanOpt
		case quartilerange:
			statOpts[quartilerange] = d.quartileRangeOpt
		case firstquartile:
			statOpts[firstquartile] = d.firstQuartileOpt
		case thirdquartile:
			statOpts[thirdquartile] = d.thirdQuartileOpt
		default:
			return nil, fmt.Errorf("Unknown statistic received %T:", stat)
		}
	}
	return statOpts, nil
}

// config options in order to calcul each required statistics once

func (d *dataBuffer) countOpt(result result) {
	result[count] = d.Count()
}

func (d *dataBuffer) sumOpt(result result) {
	result[sum] = d.Sum()
}

func (d *dataBuffer) minimumOpt(result result) {
	result[minimum] = d.Minimum()
}

func (d *dataBuffer) maximumOpt(result result) {
	result[maximum] = d.Maximum()
}

func (d *dataBuffer) rangeOpt(result result) {
	_, ok := result[minimum]
	if !ok {
		d.minimumOpt(result)
	}
	_, ok = result[maximum]
	if !ok {
		d.maximumOpt(result)
	}
	result[rangeval] = d.Range(result[minimum].(float64), result[maximum].(float64))
}

func (d *dataBuffer) meanOpt(result result) {
	_, ok := result[sum]
	if !ok {
		d.sumOpt(result)
	}
	_, ok = result[count]
	if !ok {
		d.countOpt(result)
	}
	result[mean] = d.Mean(result[sum].(float64), result[count].(int))
}

func (d *dataBuffer) medianOpt(result result) {
	result[median] = d.Median()
}

func (d *dataBuffer) modesOpt(result result) {
	result[mode] = d.Mode()
}

func (d *dataBuffer) firstQuartileOpt(result result) {
	result[firstquartile] = d.FirstQuartile()
}

func (d *dataBuffer) thirdQuartileOpt(result result) {
	result[thirdquartile] = d.ThirdQuartile()
}

func (d *dataBuffer) quartileRangeOpt(result result) {
	_, ok := result[firstquartile]
	if !ok {
		d.firstQuartileOpt(result)
	}
	_, ok = result[thirdquartile]
	if !ok {
		d.thirdQuartileOpt(result)
	}
	result[quartilerange] = d.Range(result[firstquartile].(float64), result[thirdquartile].(float64))
}

func (d *dataBuffer) varianceOpt(result result) {
	_, ok := result[mean]
	if !ok {
		d.meanOpt(result)
	}
	result[variance] = d.Variance(result[mean].(float64))
}

func (d *dataBuffer) standardDeviationOpt(result result) {
	_, ok := result[variance]
	if !ok {
		d.varianceOpt(result)
	}
	result[standarddeviation] = d.StandardDeviation(result[variance].(float64))
}

func (d *dataBuffer) secondPercentileOpt(result result) {
	result[secondpercentile], _ = d.PercentileNearestRank(2)
}

func (d *dataBuffer) ninthPercentileOpt(result result) {
	result[ninthpercentile], _ = d.PercentileNearestRank(9)
}

func (d *dataBuffer) twentyPercentileOpt(result result) {
	result[twentyfifthpercentile], _ = d.PercentileNearestRank(25)
}

func (d *dataBuffer) SeventyFifthPercentileOpt(result result) {
	result[seventyfifthpercentile], _ = d.PercentileNearestRank(75)
}

func (d *dataBuffer) ninetyFirstPercentileOpt(result result) {
	result[ninetyfirstpercentile], _ = d.PercentileNearestRank(91)
}

func (d *dataBuffer) ninetyFifthPercentileOpt(result result) {
	result[ninetyfifthpercentile], _ = d.PercentileNearestRank(95)
}

func (d *dataBuffer) ninetyEightPercentileOpt(result result) {
	result[ninetyeighthpercentile], _ = d.PercentileNearestRank(98)
}

func (d *dataBuffer) ninetyNinthPercentileOpt(result result) {
	result[ninetyninthpercentile], _ = d.PercentileNearestRank(99)
}

func (d *dataBuffer) skewnessOpt(result result) {
	_, ok := result[standarddeviation]
	if !ok {
		d.standardDeviationOpt(result)
	}
	_, ok = result[mean]
	if !ok {
		d.meanOpt(result)
	}
	result[skewness] = d.Skewness(result[mean].(float64), result[standarddeviation].(float64))
}

func (d *dataBuffer) kurtosisOpt(result result) {
	_, ok := result[standarddeviation]
	if !ok {
		d.standardDeviationOpt(result)
	}
	_, ok = result[mean]
	if !ok {
		d.meanOpt(result)
	}
	result[kurtosis] = d.Kurtosis(result[mean].(float64), result[standarddeviation].(float64))
}

func (d *dataBuffer) trimeanOpt(result result) {
	_, ok := result[thirdquartile]
	if !ok {
		d.thirdQuartileOpt(result)
	}
	_, ok = result[median]
	if !ok {
		d.medianOpt(result)
	}
	_, ok = result[firstquartile]
	if !ok {
		d.firstQuartileOpt(result)
	}
	result[trimean] = d.Trimean(result[firstquartile].(float64), result[median].(float64), result[thirdquartile].(float64))
}
