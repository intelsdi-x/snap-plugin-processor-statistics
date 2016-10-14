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
	"reflect"
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
		data := [10]float64{5, 12, 7, 9, 33, 53, 24, 16, 18, 1}

		time := [10]time.Time{time.Now().Add(12 * time.Hour),
			time.Now().Add(22 * time.Hour),
			time.Now().Add(9 * time.Hour),
			time.Now().Add(10 * time.Hour),
			time.Now().Add(1 * time.Hour),
			time.Now().Add(2 * time.Hour),
			time.Now().Add(3 * time.Hour),
			time.Now().Add(5 * time.Hour),
			time.Now().Add(6 * time.Hour),
			time.Now().Add(7 * time.Hour),
		}
		config := plugin.Config{}
		config["SlidingWindowLength"] = int64(5)

		empty := []float64{}

		Convey("Statistics for float64 data", func() {
			for i := range metrics {
				metrics[i] = plugin.Metric{
					Data:      data[i],
					Namespace: plugin.NewNamespace("foo", "bar"),
					Timestamp: time[i],
				}
			}

			statisticsObj := New()
			stats, err := statisticsObj.Process(metrics, config)

			if err != nil {
				log.Fatal(err)
			}

			modes := [][]float64{[]float64{33}, empty, empty, empty, empty, empty, empty, empty, empty, empty}

			fiveNumberSummaries := [][]float64{[]float64{33, math.NaN(), 33, math.NaN(), 33},
				[]float64{33, 33, 43, 53, 53},
				[]float64{24, 24, 33, 53, 53},
				[]float64{16, 20, 28.5, 43, 53},
				[]float64{16, 17, 24, 43, 53},
				[]float64{1, 8.5, 18, 38.5, 53},
				[]float64{1, 4, 16, 21, 24},
				[]float64{1, 4, 9, 17, 18},
				[]float64{1, 3, 7, 13.5, 18},
				[]float64{1, 3, 7, 10.5, 12},
			}

			sevenNumberSummaries := [][]float64{[]float64{33, 33, 33, 33, 33, 33, 33},
				[]float64{33, 33, 33, 43, 53, 53, 53},
				[]float64{24, 24, 24, 33, 53, 53, 53},
				[]float64{16, 16, 16, 28.5, 33, 53, 53},
				[]float64{16, 16, 18, 24, 33, 53, 53},
				[]float64{1, 1, 16, 18, 24, 53, 53},
				[]float64{1, 1, 7, 16, 18, 24, 24},
				[]float64{1, 1, 7, 9, 16, 18, 18},
				[]float64{1, 1, 5, 7, 9, 18, 18},
				[]float64{1, 1, 5, 7, 9, 12, 12},
			}
			expected := make(map[string][]float64)
			expected["Count"] = []float64{1, 2, 3, 4, 5, 5, 5, 5, 5, 5}
			expected["Mean"] = []float64{33, 43, 36.66666667, 31.5, 28.8, 22.4, 13.2, 10.2, 8, 6.8}
			expected["Median"] = []float64{33, 43, 33, 28.5, 24, 18, 16, 9, 7, 7}
			expected["Sum"] = []float64{33, 86, 110, 126, 144, 112, 66, 51, 40, 34}
			expected["Standard Deviation"] = []float64{0, 10, 12.120, 13.793, 13.467, 17.072, 8.183, 6.177, 5.657, 3.709}
			expected["Variance"] = []float64{0, 100, 146.889, 190.25, 181.36, 291.44, 66.96, 38.16, 32, 13.76}
			expected["Maximum"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["Minimum"] = []float64{33, 33, 24, 16, 16, 1, 1, 1, 1, 1}
			expected["99%-ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["95%-ile"] = []float64{33, 53, 53, 53, 53, 53, 24, 18, 18, 12}
			expected["Kurtosis"] = []float64{math.NaN(), math.NaN(), math.NaN(), 1.8964, 2.337, 2.563, 1.6874, 1.6624, 2.4383, 2.0035}
			expected["Skewness"] = []float64{math.NaN(), math.NaN(), 0.426, 0.552, 0.883, 0.744, -0.242, -0.122, 0.696, -0.195}
			expected["Trimean"] = []float64{math.NaN(), math.NaN(), 35.75, 30, 27, 20.75, 14.25, 9.75, 7.625, 6.875}
			expected["Range"] = []float64{0, 20, 29, 37, 37, 52, 23, 17, 17, 11}
			expected["Quartile_Range"] = []float64{math.NaN(), 20, 29, 23, 26, 30, 17, 13, 10.5, 7.5}

			//Tracks current location of results
			count := 0
			for i, m := range stats {
				//Captures the statistic being processed while ignoring the remaining portions of the namespace
				ns := m.Namespace.Strings()[1]

				//If all 18 statistics have been compared, then increase metric count
				if i%18 == 0 && i != 0 {
					count++
				}

				switch ns {
				case "Count":
					So(m.Data, ShouldAlmostEqual, expected["Count"][count], 0.01)
				case "Mean":
					So(m.Data, ShouldAlmostEqual, expected["Mean"][count], 0.01)
				case "Median":
					So(m.Data, ShouldAlmostEqual, expected["Median"][count], 0.01)
				case "Trimean":
					if math.IsNaN(expected["Trimean"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {

						So(m.Data, ShouldAlmostEqual, expected["Trimean"][count], 0.01)
					}
				case "Range":
					So(m.Data, ShouldAlmostEqual, expected["Range"][count], 0.01)
				case "Sum":
					So(m.Data, ShouldAlmostEqual, expected["Sum"][count], 0.01)
				case "Kurtosis":
					if math.IsNaN(expected["Kurtosis"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["Kurtosis"][count], 0.01)
					}
				case "Skewness":
					if math.IsNaN(expected["Skewness"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["Skewness"][count], 0.01)
					}
				case "Standard Deviation":
					So(m.Data, ShouldAlmostEqual, expected["Standard Deviation"][count], 0.01)
				case "Variance":
					So(m.Data, ShouldAlmostEqual, expected["Variance"][count], 0.01)
				case "Maximum":
					So(m.Data, ShouldAlmostEqual, expected["Maximum"][count], 0.01)
				case "Minimum":
					So(m.Data, ShouldAlmostEqual, expected["Minimum"][count], 0.01)
				case "99%-ile":
					So(m.Data, ShouldAlmostEqual, expected["99%-ile"][count], 0.01)
				case "95%-ile":
					So(m.Data, ShouldAlmostEqual, expected["95%-ile"][count], 0.01)
				case "Mode":
					So(m.Data, ShouldResemble, modes[count])
				case "Five_Number_Summary":
					switch reflect.TypeOf(m.Data).Kind() {
					case reflect.Slice:
						sli := reflect.ValueOf(m.Data)
						for j, v := range fiveNumberSummaries[count] {
							if math.IsNaN(v) {
								So(sli.Index(j).Interface(), ShouldNotBeNil)
							} else {
								So(sli.Index(j).Interface(), ShouldAlmostEqual, v, 0.01)
							}
						}
					}
				case "Seven_Number_Summary":
					switch reflect.TypeOf(m.Data).Kind() {
					case reflect.Slice:
						sl := reflect.ValueOf(m.Data)
						for j, v := range sevenNumberSummaries[count] {
							if math.IsNaN(v) {
								So(sl.Index(j).Interface(), ShouldNotBeNil)
							} else {
								So(sl.Index(j).Interface(), ShouldAlmostEqual, v, 0.01)
							}
						}
					}
				case "Quartile_Range":
					if math.IsNaN(expected["Quartile_Range"][count]) {
						So(m.Data, ShouldNotBeNil)
					} else {
						So(m.Data, ShouldAlmostEqual, expected["Quartile_Range"][count], 0.01)
					}
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

				data := "I am an unknow data Type"
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
