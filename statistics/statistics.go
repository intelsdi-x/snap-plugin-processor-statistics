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
	"errors"
	"fmt"
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type Plugin struct {
	buffer map[string]*dataBuffer
}

const (
	Name    = "statistics"
	Version = 2
)

var (
	statList = []string{"count", "mean", "median", "standard_deviation", "variance", "95%_ile", "99%_ile", "2%_ile", "9%_ile", "25%_ile", "75%_ile", "91%_ile", "98%_ile", "minimum", "maximum", "range", "mode", "kurtosis", "skewness", "sum", "trimean", "quartile_range"}
)

// New() returns a new instance of this
func New() *Plugin {
	buffer := make(map[string]*dataBuffer)
	p := &Plugin{buffer: buffer}
	return p
}

// GetConfigPolicy returns the config policy
func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewIntRule([]string{""}, "slidingWindowLength", false, plugin.SetDefaultInt(100), plugin.SetMinInt(1))
	policy.AddNewIntRule([]string{""}, "slidingFactor", false, plugin.SetDefaultInt(100), plugin.SetMinInt(1))
	policy.AddNewStringRule([]string{""}, "statistics", false, plugin.SetDefaultString(strings.Join(statList, ",")))
	return *policy, nil

}

func GetConfig(cfg plugin.Config) (slidingWinLen, slidingFac int, statistics []string, err error) {
	var stats string
	stats, err = cfg.GetString("statistics")
	if err != nil {
		return
	}
	statistics = strings.Split(stats, ",")
	var tmp int64

	tmp, err = cfg.GetInt("SlidingWindowLength")
	if err != nil {
		return
	}
	slidingWinLen = int(tmp)
	tmp, err = cfg.GetInt("slidingFactor")
	if err != nil {
		return
	}
	slidingFac = int(tmp)

	errString := ""

	if slidingFac > slidingWinLen {
		errString += "Sliding Factor is greater than window length and it shouldn't be\n"
	}

	if errString != "" {
		err = fmt.Errorf(errString)
	}
	return
}

func dataToFloat64(data interface{}) (float64, error) {
	var buffer float64
	switch v := data.(type) {
	case int:
		buffer = float64(data.(int))
	case int32:
		buffer = float64(data.(int32))
	case int64:
		buffer = float64(data.(int64))
	case float64:
		buffer = data.(float64)
	case float32:
		buffer = float64(data.(float32))
	case uint64:
		buffer = float64(data.(uint64))
	case uint32:
		buffer = float64(data.(uint32))
	default:
		st := fmt.Sprintf("Unknown data received in calculateStats(): Type %T", v)
		return 0, errors.New(st)
	}
	return buffer, nil
}

func (p *Plugin) Process(metrics []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	var result []plugin.Metric
	for _, metric := range metrics {
		slidingWindowLength, slidingFactor, stats, err := GetConfig(cfg)

		floatValue, err := dataToFloat64(metric.Data)
		// convert any number to float64

		if err != nil {
			return nil, err
		}
		nsSlice := metric.Namespace.Strings()
		ns := strings.Join(nsSlice, "")
		buffer, ok := p.buffer[ns]
		if !ok {
			//buffer doesn't exist, so create one
			p.buffer[ns] = &dataBuffer{
				data:            make([]*data, 0, slidingWindowLength),
				dataByTimestamp: make([]*data, 0, slidingWindowLength)}
		} else {
			if slidingWindowLength != cap(buffer.data) {
				// TODO: test if buffer size from the config is different than cap(p.buffer[ns])
			}
		}

		buffer.Insert(floatValue, metric.Timestamp)
		if buffer.slidingFactorIndex%slidingFactor == 0 {
			// add a new element to the sorted list
			mts, err := buffer.GetStats(stats, nsSlice)
			if err != nil {
				return nil, err
			}
			result = append(result, mts...)
		}
		buffer.slidingFactorIndex++
	}
	return result, nil
}
