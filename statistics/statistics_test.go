//
// +build small

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
	"log"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStatisticsProcessor(t *testing.T) {
	Convey("Meta should return metadata for the plugin", t, func() {
		Convey("So Name should equal statistics", func() {
			So(Name, ShouldEqual, "statistics")
		})
		Convey("So Version should equal version", func() {
			So(Version, ShouldEqual, Version)
		})
	})

	proc := New()
	Convey("Create statistics processor", t, func() {
		Convey("So proc should not be nil", func() {
			So(proc, ShouldNotBeNil)
		})
		Convey("So proc should be of type statisticsProcessor", func() {
			So(proc, ShouldHaveSameTypeAs, &Plugin{})
		})
		Convey("proc.GetConfigPolicy should return a config policy", func() {
			configPolicy, err := proc.GetConfigPolicy()
			Convey("So config policy should be a plugin.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, plugin.ConfigPolicy{})
			})
			Convey("So err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestStatisticsProcessorMetrics(t *testing.T) {
	Convey("Statistics Processor tests", t, func() {
		metrics := make([]plugin.Metric, 10)
		var stats []plugin.Metric
		var err error
		data := []float64{33, 53, 24, 16, 18, 1, 7, 9, 5, 12}
		time := [10]time.Time{time.Now().Add(22 * time.Hour),
			time.Now().Add(12 * time.Hour),
			time.Now().Add(10 * time.Hour),
			time.Now().Add(9 * time.Hour),
			time.Now().Add(7 * time.Hour),
			time.Now().Add(6 * time.Hour),
			time.Now().Add(5 * time.Hour),
			time.Now().Add(3 * time.Hour),
			time.Now().Add(2 * time.Hour),
			time.Now().Add(1 * time.Hour),
		}

		config := plugin.Config{}
		config["slidingWindowLength"] = int64(5)
		config["slidingFactor"] = int64(1)
		config["statistics"] = strings.Join(statList, ",")

		empty := []float64{}
		modes := [][]float64{[]float64{33}, empty, empty, empty, empty, empty, empty, empty, empty, empty}

		expected := make(map[string][]float64)
		expected[sum] = []float64{33, 86, 110, 126, 144, 112, 66, 51, 40, 34}
		expected[count] = []float64{1, 2, 3, 4, 5, 5, 5, 5, 5, 5}
		expected[minimum] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
		expected[maximum] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
		expected[mean] = []float64{33, 43, 36.66666667, 31.5, 28.8, 22.4, 13.2, 10.2, 8, 6.8}
		expected[median] = []float64{33, 43, 33, 28.5, 24, 18, 16, 9, 7, 7}
		expected[standarddeviation] = []float64{0, 10, 12.120, 13.793, 13.467, 17.072, 8.183, 6.177, 5.657, 3.709}
		expected[variance] = []float64{0, 100, 146.889, 190.25, 181.36, 291.44, 66.96, 38.16, 32, 13.76}
		expected[ninetyninthpercentile] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
		expected[ninetyfifthpercentile] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
		expected[secondpercentile] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
		expected[ninthpercentile] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
		expected[twentyfifthpercentile] = []float64{33, 33, 24, 16, 18, 16, 7, 7, 5, 5}
		expected[seventyfifthpercentile] = []float64{33, 53, 53, 33, 33, 24, 18, 16, 9, 9}
		expected[ninetyfirstpercentile] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
		expected[ninetyeighthpercentile] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
		expected[kurtosis] = []float64{math.NaN(), math.NaN(), math.NaN(), 1.8964, 2.337, 2.563, 1.6874, 1.6624, 2.4383, 2.0035}
		expected[skewness] = []float64{math.NaN(), math.NaN(), 0.426, 0.552, 0.883, 0.744, -0.242, -0.122, 0.696, -0.195}
		expected[trimean] = []float64{math.NaN(), math.NaN(), 33.25, 30, 24.5, 17.125, 13.5, 9.5, 6.5, 6.5}
		expected[rangeval] = []float64{0, 20, 29, 37, 37, 52, 23, 17, 17, 11}
		expected[quartilerange] = []float64{0, 15, 19, 23, 16, 15.5, 14, 12, 6, 6}
		expected[firstquartile] = []float64{33, 33, 24, 20, 17, 8.5, 4, 4, 3, 3}
		expected[thirdquartile] = []float64{33, 48, 43, 43, 33, 24, 18, 16, 9, 9}

		Convey("Statistics for float64 data", func() {
			statisticsObj := New()
			for i := range metrics {
				mts := []plugin.Metric{plugin.Metric{
					Data:      data[i],
					Namespace: plugin.NewNamespace("foo", "bar"),
					Timestamp: time[i],
					Config:    config,
				}}
				stats, err = statisticsObj.Process(mts, config)
				if err != nil {
					log.Fatal(err)
				}

				//Tracks current location of results
				var ns string
				for _, m := range stats {
					//Captures the statistic being processed while ignoring the remaining portions of the namespace
					dynamic, _ := m.Namespace.IsDynamic()
					nsSlice := m.Namespace.Strings()
					if dynamic {
						ns = nsSlice[len(nsSlice)-2]
					} else {
						ns = nsSlice[len(nsSlice)-1]
					}
					log.Printf("\n%v:%v", i, ns)

					switch ns {
					case count:
						So(m.Data, ShouldAlmostEqual, expected[count][i], 0.01)
					case mean:
						So(m.Data, ShouldAlmostEqual, expected[mean][i], 0.01)
					case median:
						So(m.Data, ShouldAlmostEqual, expected[median][i], 0.01)
					case trimean:
						if math.IsNaN(expected[trimean][i]) {
							So(m.Data, ShouldNotBeNil)
						} else {

							So(m.Data, ShouldAlmostEqual, expected[trimean][i], 0.01)
						}
					case rangeval:
						So(m.Data, ShouldAlmostEqual, expected[rangeval][i], 0.01)
					case sum:
						So(m.Data, ShouldAlmostEqual, expected[sum][i], 0.01)
					case kurtosis:
						if math.IsNaN(expected[kurtosis][i]) {
							So(m.Data, ShouldNotBeNil)
						} else {
							So(m.Data, ShouldAlmostEqual, expected[kurtosis][i], 0.01)
						}
					case skewness:
						if math.IsNaN(expected[skewness][i]) {
							So(m.Data, ShouldNotBeNil)
						} else {
							So(m.Data, ShouldAlmostEqual, expected[skewness][i], 0.01)
						}
					case standarddeviation:
						So(m.Data, ShouldAlmostEqual, expected[standarddeviation][i], 0.01)
					case variance:
						So(m.Data, ShouldAlmostEqual, expected[variance][i], 0.01)
					case maximum:
						So(m.Data, ShouldAlmostEqual, expected[maximum][i], 0.01)
					case minimum:
						So(m.Data, ShouldAlmostEqual, expected[minimum][i], 0.01)
					case ninetyninthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[ninetyninthpercentile][i], 0.01)
					case ninetyfifthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[ninetyfifthpercentile][i], 0.01)
					case secondpercentile:
						So(m.Data, ShouldAlmostEqual, expected[secondpercentile][i], 0.01)
					case ninthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[ninthpercentile][i], 0.01)
					case twentyfifthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[twentyfifthpercentile][i], 0.01)
					case seventyfifthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[seventyfifthpercentile][i], 0.01)
					case ninetyfirstpercentile:
						So(m.Data, ShouldAlmostEqual, expected[ninetyfirstpercentile][i], 0.01)
					case ninetyeighthpercentile:
						So(m.Data, ShouldAlmostEqual, expected[ninetyeighthpercentile][i], 0.01)
					case mode:
						So(m.Data, ShouldResemble, modes[i])
					case quartilerange:
						if math.IsNaN(expected[quartilerange][i]) {
							So(m.Data, ShouldNotBeNil)
						} else {
							So(m.Data, ShouldAlmostEqual, expected[quartilerange][i], 0.01)
						}
					case firstquartile:
						So(m.Data, ShouldAlmostEqual, expected[firstquartile][i], 0.01)
					case thirdquartile:
						So(m.Data, ShouldAlmostEqual, expected[thirdquartile][i], 0.01)
					default:
						log.Println("Raw metric found")
						log.Printf("Data: %v", ns)
					}
				}
			}

			var metricsNew []plugin.Metric
			So(metrics, ShouldNotResemble, metricsNew)
		})

		Convey("Statistics for unknown data type", func() {
			for i := range metrics {

				data := "I am an unknown data Type"
				metrics[i] = plugin.Metric{
					Data:      data,
					Namespace: plugin.NewNamespace("foo", "bar"),
					Timestamp: time[i],
				}
			}

			var metricsNew []plugin.Metric
			statisticObj := New()
			receivedData, _ := statisticObj.Process(metricsNew, config)

			So(metrics, ShouldNotResemble, receivedData)
		})

	})
}
