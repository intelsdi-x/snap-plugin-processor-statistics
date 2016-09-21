//
// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2015 Intel Corporation

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
	"bytes"
	"encoding/gob"
	"log"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

//Random number generator
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func TestStatisticsProcessor(t *testing.T) {
	meta := Meta()
	Convey("Meta should return metadata for the plugin", t, func() {
		Convey("So meta.Name should equal statistics", func() {
			So(meta.Name, ShouldEqual, "statistics")
		})
		Convey("So meta.Version should equal version", func() {
			So(meta.Version, ShouldEqual, version)
		})
		Convey("So meta.Type should be of type plugin.ProcessorPluginType", func() {
			So(meta.Type, ShouldResemble, plugin.ProcessorPluginType)
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
			configPolicy, _ := proc.GetConfigPolicy()
			Convey("So config policy should be a cpolicy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, &cpolicy.ConfigPolicy{})
			})
			testConfig := make(map[string]ctypes.ConfigValue)
			testConfig["SlidingWindowLength"] = ctypes.ConfigValueInt{Value: 23}
			cfg, errs := configPolicy.Get([]string{""}).Process(testConfig)
			Convey("So config policy should process testConfig and return a config", func() {
				So(cfg, ShouldNotBeNil)
			})
			Convey("So testConfig processing should return no errors", func() {
				So(errs.HasErrors(), ShouldBeFalse)
			})
		})
	})
}

func TestStatisticsProcessorMetrics(t *testing.T) {
	Convey("Statistics Processor tests", t, func() {
		metrics := make([]plugin.MetricType, 10)
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
		config := make(map[string]ctypes.ConfigValue)
		config["SlidingWindowLength"] = ctypes.ConfigValueInt{Value: 5}

		empty := []float64(nil)

		Convey("Statistics for float64 data", func() {
			for i := range metrics {
				metrics[i] = plugin.MetricType{
					Data_:      data[i],
					Namespace_: core.NewNamespace("foo", "bar"),
					Timestamp_: time[i],
				}
			}

			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			statisticsObj := New()
			_, stats, err := statisticsObj.Process("snap.gob", buf.Bytes(), config)

			if err != nil {
				log.Fatal(err)
			}

			var results []plugin.MetricType
			dec := gob.NewDecoder(bytes.NewBuffer(stats))
			err = dec.Decode(&results)

			if err != nil {
				log.Fatal("decode", err)
			}

			modes := [][]float64{[]float64{33}, empty, empty, empty, empty, empty, empty, empty, empty, empty}

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

			//Tracks current location of results
			count := 0
			for i, m := range results {
				//Captures the statistic being processed while ignoring the remaining portions of the namespace
				ns := m.Namespace().Strings()[1]

				//If all 15 statistics have been compared, then increase metric count
				if i%15 == 0 && i != 0 {
					count++
				}

				switch ns {
				case "Count":
					So(m.Data(), ShouldAlmostEqual, expected["Count"][count], 0.01)
				case "Mean":
					So(m.Data(), ShouldAlmostEqual, expected["Mean"][count], 0.01)
				case "Median":
					So(m.Data(), ShouldAlmostEqual, expected["Median"][count], 0.01)
				case "Trimean":
					if math.IsNaN(expected["Trimean"][count]) {
						So(m.Data(), ShouldNotBeNil)
					} else {

						So(m.Data(), ShouldAlmostEqual, expected["Trimean"][count], 0.01)
					}
				case "Range":
					So(m.Data(), ShouldAlmostEqual, expected["Range"][count], 0.01)
				case "Sum":
					So(m.Data(), ShouldAlmostEqual, expected["Sum"][count], 0.01)
				case "Kurtosis":
					if math.IsNaN(expected["Kurtosis"][count]) {
						So(m.Data(), ShouldNotBeNil)
					} else {
						So(m.Data(), ShouldAlmostEqual, expected["Kurtosis"][count], 0.01)
					}
				case "Skewness":
					if math.IsNaN(expected["Skewness"][count]) {
						So(m.Data(), ShouldNotBeNil)
					} else {
						So(m.Data(), ShouldAlmostEqual, expected["Skewness"][count], 0.01)
					}
				case "Standard Deviation":
					So(m.Data(), ShouldAlmostEqual, expected["Standard Deviation"][count], 0.01)
				case "Variance":
					So(m.Data(), ShouldAlmostEqual, expected["Variance"][count], 0.01)
				case "Maximum":
					So(m.Data(), ShouldAlmostEqual, expected["Maximum"][count], 0.01)
				case "Minimum":
					So(m.Data(), ShouldAlmostEqual, expected["Minimum"][count], 0.01)
				case "99%-ile":
					So(m.Data(), ShouldAlmostEqual, expected["99%-ile"][count], 0.01)
				case "95%-ile":
					So(m.Data(), ShouldAlmostEqual, expected["95%-ile"][count], 0.01)
				case "Mode":
					So(m.Data(), ShouldResemble, modes[count])
				default:
					log.Println("Raw metric found")
					log.Println("Data: %v", ns)
				}
			}

			var metricsNew []plugin.MetricType
			So(metrics, ShouldNotResemble, metricsNew)
		})

		Convey("Statistics for unknown data type", func() {
			for i := range metrics {

				data := "I am an unknow data Type"
				metrics[i] = plugin.MetricType{
					Data_:      data,
					Namespace_: core.NewNamespace("foo", "bar"),
					Timestamp_: time[i],
				}
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			statisticObj := New()
			_, receivedData, _ := statisticObj.Process("snap.gob", buf.Bytes(), config)

			var metricsNew []plugin.MetricType

			//Decodes the content into MetricType
			dec := gob.NewDecoder(bytes.NewBuffer(receivedData))
			dec.Decode(&metricsNew)
			So(metrics, ShouldNotResemble, metricsNew)
		})

	})
}
