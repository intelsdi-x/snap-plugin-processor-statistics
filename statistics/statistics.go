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

package movingaverage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	pluginName = "statistics"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

var stats = []string{"mean", "median", "stddev", "variance", "90%-ile", "95%-ile", "99%-ile", "99.99%-ile",
	"smean", "smedian", "sstddev", "svariance", "s90%-ile", "s95%-ile", "s99%-ile", "s99.99%-ile"}

type Plugin struct {
	slidingWinSize int
}

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		pluginName,
		version,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType})
}

func New() *Plugin {
	p := &Plugin{slidingWinSize: 30}
	return p
}

func mean(m plugin.MetricType, logger *log.Logger) (float64, error) {

	switch t := m.Data().(type) {
	default:
		logger.Printf("Unknown data received: Type %T", t)
		return 0.0, errors.New("Unknown data received: Type")
	case float64:
		if len(m.Data()) == 0 {
			return math.NaN(), errors.New("No Data")
		}

		sum, _ := m.Data().Sum()

		return sum / float64(len(m.Data())), nil
	}

}

func (p *Plugin) calculateStats(m plugin.MetricType, logger *log.Logger) (map[string]float64, error) {
	result := make(map[string]float64, 16)

	val, err := mean(m)
	if err != nil {
		return nil, err
	}

	result["mean"] = val

	//result["median"] = percentile(m, 50)

	//	switch v := m.Data().(type) {
	//	default:
	//		logger.Printf("Unknown data received: Type %T", v)
	//		return 0.0, errors.New("Unknown data received: Type")
	//	case float64:
	//		result["mean"] = mean()

	//	}

	return result, nil
}

func (p *Plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewIntegerRule("SlidingWindowLength", true)
	if err != nil {
		return nil, err
	}

	r1.Description = "Length for sliding window"
	config.Add(r1)
	cp.Add([]string{""}, config)

	return cp, nil
}

func (p *Plugin) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	logger := log.New()
	logger.Println("Statistics Processor started")

	var metrics []plugin.MetricType

	if config != nil {
		if config["SlidingWindowLength"].(ctypes.ConfigValueInt).Value > 0 {
			p.slidingWinSize = config["SlidingWindowLength"].(ctypes.ConfigValueInt).Value
		} else {
			p.slidingWinSize = 30
		}
	} else {
		p.slidingWinSize = 30
	}

	//Decodes the content into MetricType
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for i, m := range metrics {
		//Determining the type of data
		logger.Printf("Data received %v", metrics[i].Data())

		metrics[i].Data_, _ = p.calculateStats(m, logger)
		logger.Printf("Statistics %v", metrics[i].Data())

	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)

	return contentType, buf.Bytes(), nil
}
