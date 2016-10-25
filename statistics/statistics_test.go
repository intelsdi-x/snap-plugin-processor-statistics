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
		//Sorted data : 1, 5, 7, 9, 12, 16, 18, 24, 33, 53
		time := [10]time.Time{time.Now().Add(1 * time.Hour),
			time.Now().Add(2 * time.Hour),
			time.Now().Add(3 * time.Hour),
			time.Now().Add(5 * time.Hour),
			time.Now().Add(6 * time.Hour),
			time.Now().Add(7 * time.Hour),
			time.Now().Add(9 * time.Hour),
			time.Now().Add(10 * time.Hour),
			time.Now().Add(12 * time.Hour),
			time.Now().Add(22 * time.Hour),
		}

		// time := [10]time.Time{time.Now().Add(12 * time.Hour),
		// 	time.Now().Add(22 * time.Hour),
		// 	time.Now().Add(9 * time.Hour),
		// 	time.Now().Add(10 * time.Hour),
		// 	time.Now().Add(1 * time.Hour),
		// 	time.Now().Add(2 * time.Hour),
		// 	time.Now().Add(3 * time.Hour),
		// 	time.Now().Add(5 * time.Hour),
		// 	time.Now().Add(6 * time.Hour),
		// 	time.Now().Add(7 * time.Hour),
		// }
		config := plugin.Config{}
		config["slidingWindowLength"] = int64(5)
		config["slidingFactor"] = int64(1)
		config["statistics"] = strings.Join(statList, ",")

		empty := []float64{}

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

			}
			//Sorted data : 1, 5, 7, 9, 12, 16, 18, 24, 33, 53
			modes := [][]float64{[]float64{33}, empty, empty, empty, empty, empty, empty, empty, empty, empty}

			expected := make(map[string][]float64)
			//MAIN: expected["count"] = []float64{1, 2, 3, 4, 5, 5, 5, 5, 5, 5}
			//MAIN: expected["maximum"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			//MAIN: expected["minimum"] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
			//MAIN: expected["mean"] = []float64{33, 43, 36.66666667, 31.5, 28.8, 22.4, 13.2, 10.2, 8, 6.8}
			//MAIN: expected["median"] = []float64{33, 43, 33, 28.5, 24, 18, 16, 9, 7, 7}
			//MAIN: expected["sum"] = []float64{33, 86, 110, 126, 144, 112, 66, 51, 40, 34}
			expected["count"] = []float64{1, 2, 3, 4, 5, 5, 5, 5, 5, 5}
			expected["minimum"] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
			expected["maximum"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["mean"] = []float64{1, 3, 4.33, 5.5, 6.8, 9.8, 12.4, 15.8, 20.6, 28.8}
			expected["median"] = []float64{1, 3, 5, 6, 7, 9, 12, 16, 18, 24}
			expected["sum"] = []float64{1, 6, 13, 22, 34, 49, 62, 79, 103, 144}
			expected["standard_deviation"] = []float64{0, 10, 12.120, 13.793, 13.467, 17.072, 8.183, 6.177, 5.657, 3.709}
			expected["variance"] = []float64{0, 100, 146.889, 190.25, 181.36, 291.44, 66.96, 38.16, 32, 13.76}
			expected["99%_ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["95%_ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["2%_ile"] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
			expected["9%_ile"] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
			expected["25%_ile"] = []float64{33, 33, 24, 16, 18, 16, 7, 7, 5, 5}
			expected["75%_ile"] = []float64{33, 53, 53, 33, 33, 24, 18, 16, 9, 9}
			expected["91%_ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["98%_ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["kurtosis"] = []float64{math.NaN(), math.NaN(), math.NaN(), 1.8964, 2.337, 2.563, 1.6874, 1.6624, 2.4383, 2.0035}
			expected["skewness"] = []float64{math.NaN(), math.NaN(), 0.426, 0.552, 0.883, 0.744, -0.242, -0.122, 0.696, -0.195}
			expected["trimean"] = []float64{math.NaN(), math.NaN(), 35.75, 30, 27, 20.75, 14.25, 9.75, 7.625, 6.875}
			expected["range"] = []float64{0, 20, 29, 37, 37, 52, 23, 17, 17, 11}
			expected["quartile_range"] = []float64{math.NaN(), 20, 29, 23, 26, 30, 17, 13, 10.5, 7.5}
			expected["first_quartile"] = []float64{33, 33, 24, 16, 18, 16, 7, 7, 5, 5}
			expected["third_quartile"] = []float64{33, 53, 53, 33, 33, 24, 18, 16, 9, 9}

			//Tracks current location of results
			count := 0
			var ns string
			for i, m := range stats {
				//Captures the statistic being processed while ignoring the remaining portions of the namespace
				dynamic, _ := m.Namespace.IsDynamic()
				nsSlice := m.Namespace.Strings()
				if dynamic {
					ns = nsSlice[len(nsSlice)-2]
				} else {
					ns = nsSlice[len(nsSlice)-1]
				}

				//If all 22 statistics have been compared, then increase metric count
				if i%22 == 0 && i != 0 {
					count++
				}

				switch ns {
				case "count":
					So(m.Data, ShouldAlmostEqual, expected["count"][count], 0.01)
				case "mean":
					So(m.Data, ShouldAlmostEqual, expected["mean"][count], 0.01)
				case "median":
					So(m.Data, ShouldAlmostEqual, expected["median"][count], 0.01)
				case "trimean":
					if math.IsNaN(expected["trimean"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {

						So(m.Data, ShouldAlmostEqual, expected["trimean"][count], 0.01)
					}
				case "range":
					So(m.Data, ShouldAlmostEqual, expected["range"][count], 0.01)
				case "sum":
					So(m.Data, ShouldAlmostEqual, expected["sum"][count], 0.01)
				case "kurtosis":
					if math.IsNaN(expected["kurtosis"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["kurtosis"][count], 0.01)
					}
				case "skewness":
					if math.IsNaN(expected["skewness"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["skewness"][count], 0.01)
					}
				case "standard_deviation":
					So(m.Data, ShouldAlmostEqual, expected["standard_deviation"][count], 0.01)
				case "variance":
					So(m.Data, ShouldAlmostEqual, expected["variance"][count], 0.01)
				case "maximum":
					So(m.Data, ShouldAlmostEqual, expected["maximum"][count], 0.01)
				case "minimum":
					So(m.Data, ShouldAlmostEqual, expected["minimum"][count], 0.01)
				case "99%_ile":
					So(m.Data, ShouldAlmostEqual, expected["99%_ile"][count], 0.01)
				case "95%_ile":
					So(m.Data, ShouldAlmostEqual, expected["95%_ile"][count], 0.01)
				case "2%_ile":
					So(m.Data, ShouldAlmostEqual, expected["2%_ile"][count], 0.01)
				case "9%_ile":
					So(m.Data, ShouldAlmostEqual, expected["9%_ile"][count], 0.01)
				case "25%_ile":
					So(m.Data, ShouldAlmostEqual, expected["25%_ile"][count], 0.01)
				case "75%_ile":
					So(m.Data, ShouldAlmostEqual, expected["75%_ile"][count], 0.01)
				case "91%_ile":
					So(m.Data, ShouldAlmostEqual, expected["91%_ile"][count], 0.01)
				case "98%_ile":
					So(m.Data, ShouldAlmostEqual, expected["98%_ile"][count], 0.01)
				case "mode":
					So(m.Data, ShouldResemble, modes[count])
				case "quartile_range":
					if math.IsNaN(expected["quartile_range"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["quartile_range"][count], 0.01)
					}
				case "first_quartile":
					So(m.Data, ShouldAlmostEqual, expected["first_quartile"][count], 0.01)
				case "third_quartile":
					So(m.Data, ShouldAlmostEqual, expected["third_quartile"][count], 0.01)
				default:
					log.Println("Raw metric found")
					log.Println("Data: %v", ns)
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
