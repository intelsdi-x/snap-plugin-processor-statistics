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
	Version = 3
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
	policy.AddNewIntRule([]string{""}, "slidingFactor", false, plugin.SetDefaultInt(1), plugin.SetMinInt(1))
	policy.AddNewStringRule([]string{""}, "statistics", false, plugin.SetDefaultString(strings.Join(statList, ",")))
	return *policy, nil

}

// Process processes the data, inputs the data into sorted buffer and calls the GetStats method
func (p *Plugin) Process(metrics []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	var result []plugin.Metric
	slidingWindowLength, slidingFactor, stats, err := GetConfig(cfg)
	if err != nil {
		return nil, err
	}

	for _, metric := range metrics {
		// convert any number to float64
		floatValue, err := dataToFloat64(metric.Data)
		if err != nil {
			return nil, err
		}

		nsSlice := metric.Namespace.Strings()
		ns := strings.Join(nsSlice, "")
		_, ok := p.buffer[ns]
		if !ok {
			//if there is no buffer for this particular namespace, then we create a new one
			p.buffer[ns] = &dataBuffer{
				data:            make([]*data, 0, slidingWindowLength),
				dataByTimestamp: make([]*data, 0, slidingWindowLength)}

		} else {
			if slidingWindowLength != cap(p.buffer[ns].data) {
				// TODO: test if buffer size from the config is different than cap(p.buffer[ns])
			}
		}

		p.buffer[ns].Insert(floatValue, metric.Timestamp)
		// add a new element to the sorted list
		if p.buffer[ns].slidingFactorIndex%slidingFactor == 0 {
			mts, err := p.buffer[ns].GetStats(stats, nsSlice)
			if err != nil {
				return nil, err
			}
			result = append(result, mts...)
		}
		p.buffer[ns].slidingFactorIndex++
	}
	return result, nil
}

// GetConfig returns the config policy
func GetConfig(cfg plugin.Config) (slidingWinLen, slidingFac int, statistics []string, err error) {
	var stats string
	stats, err = cfg.GetString("statistics")
	if err != nil {
		err = fmt.Errorf("\"statistics\": %v", err)
		return
	}
	statistics = strings.Split(stats, ",")
	var tmp int64

	tmp, err = cfg.GetInt("slidingWindowLength")
	if err != nil {
		err = fmt.Errorf("\"slidingwindowlength\": %v", err)
		return
	}
	slidingWinLen = int(tmp)
	tmp, err = cfg.GetInt("slidingFactor")
	if err != nil {
		err = fmt.Errorf("\"slidingfactor\": %v", err)
		return
	}
	slidingFac = int(tmp)

	if slidingFac > slidingWinLen {
		err = fmt.Errorf("Sliding Factor is greater than window length and it shouldn't be")
	}
	return
}

// converts data to float64 type
func dataToFloat64(data interface{}) (float64, error) {
	var value float64
	if data == nil {
		e := fmt.Sprintf("Data is empty : Type %T", data)
		return 0, errors.New(e)
	} else if data != nil {
		switch v := data.(type) {
		case int:
			value = float64(data.(int))
		case int32:
			value = float64(data.(int32))
		case int64:
			value = float64(data.(int64))
		case float64:
			value = data.(float64)
		case float32:
			value = float64(data.(float32))
		case uint64:
			value = float64(data.(uint64))
		case uint32:
			value = float64(data.(uint32))
		default:
			st := fmt.Sprintf("Unknown data received in calculateStats(): Type %T", v)
			return 0, errors.New(st)
		}
	}
	return value, nil
}
