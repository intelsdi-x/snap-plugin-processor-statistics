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
	"bytes"
	"encoding/gob"
	"errors"
	"math"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/montanaflynn/stats"
)

const (
	pluginName = "statistics"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

type Plugin struct {
	buffer        map[string][]interface{}
	bufferMaxSize int
	bufferCurSize int
	bufferIndex   int
	logger        *log.Logger
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
	buffer := make(map[string][]interface{})
	p := &Plugin{buffer: buffer,
		bufferMaxSize: 100,
		bufferCurSize: 0,
		bufferIndex:   0,
		logger:        log.New()}
	return p
}

func (p *Plugin) calculateStats(buff interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var buffer []float64

	p.logger.Printf("Buff %v", buff)

	var valRange [2]float64 //stores range values
	for _, val := range buff.([]interface{}) {
		switch v := val.(type) {
		default:
			p.logger.Printf("Unknown data received: Type %T", v)

			return nil, errors.New("Unknown data received: Type")
		case int:
			buffer = append(buffer, float64(val.(int)))
		case int32:
			buffer = append(buffer, float64(val.(int32)))
		case int64:
			buffer = append(buffer, float64(val.(int64)))
		case float64:
			buffer = append(buffer, val.(float64))
		case float32:
			buffer = append(buffer, float64(val.(float32)))
		case uint64:
			buffer = append(buffer, float64(val.(uint64)))
		case uint32:
			buffer = append(buffer, float64(val.(uint32)))
		}
	}

	p.logger.Printf("Buffer %v", buffer)

	result["Count"] = len(buffer)

	val, err := stats.Mean(buffer)
	if err != nil {
		return nil, err
	}

	result["mean"] = val

	val, err = stats.Median(buffer)
	if err != nil {
		return nil, err
	}

	result["median"] = val

	val, err = stats.StandardDeviation(buffer)
	if err != nil {
		return nil, err
	}

	result["stddev"] = val

	val, err = stats.Variance(buffer)
	if err != nil {
		return nil, err
	}

	result["var"] = val

	val, err = stats.Percentile(buffer, 95)
	if err != nil {
		return nil, err
	}

	result["95%-ile"] = val

	val, err = stats.Percentile(buffer, 99)
	if err != nil {
		return nil, err
	}

	result["99%-ile"] = val

	val, err = stats.Min(buffer)
	if err != nil {
		return nil, err
	}

	result["Kurtosis"] = p.Kurtosis(buffer)

	result["Minimum"] = val
	valRange[0] = val

	val, err = stats.Max(buffer)
	if err != nil {
		return nil, err
	}

	result["Maximum"] = val
	valRange[1] = val

	result["Range"] = valRange

	var valArr []float64
	valArr, err = stats.Mode(buffer)
	if err != nil {
		return nil, err
	}

	result["Mode"] = valArr

	val, err = stats.Sum(buffer)
	if err != nil {
		return nil, err
	}

	result["Skewness"] = p.Skewness(buffer)

	result["Sum"] = val

	val, err = stats.Trimean(buffer)
	if err != nil {
		return nil, err
	}

	result["Trimean"] = val

	return result, nil
}

//Calculates the population skewness from buffer
func (p *Plugin) Skewness(buffer []float64) float64 {
	if len(buffer) == 0 {
		return 0
	}
	var num float64
	var den float64
	var mean float64

	mean, err := stats.Mean(buffer)
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range buffer {
		num += math.Pow(val-mean, 3)
		den += math.Pow(val-mean, 2)
	}

	return math.Sqrt(float64(len(buffer))) * num / math.Pow(den, 3/2)

}

//Calculates the population kurtosis from buffer
func (p *Plugin) Kurtosis(buffer []float64) float64 {
	if len(buffer) == 0 {
		return 0
	}
	var num float64
	var den float64
	var mean float64

	mean, err := stats.Mean(buffer)
	if err != nil {
		log.Fatal(err)
	}

	for _, val := range buffer {
		num += math.Pow(val-mean, 4)
		den += math.Pow(val-mean, 2)
	}
	return float64(len(buffer)) * num / math.Pow(den, 2)
}

func concatNameSpace(namespace []string) string {
	completeNamespace := ""
	for _, ns := range namespace {
		completeNamespace += ns
	}
	return completeNamespace
}

func (p *Plugin) insertInToBuffer(val interface{}, ns []string) {

	if p.bufferCurSize == 0 {
		p.logger.Printf("In if statement with buffer max size %v", p.bufferMaxSize)
		var buff = make([]interface{}, p.bufferMaxSize)
		buff[0] = val
		p.buffer[concatNameSpace(ns)] = buff
	} else {
		p.logger.Printf("In else statement with index %v", p.bufferIndex)
		p.buffer[concatNameSpace(ns)][p.bufferIndex] = val
	}
}

func (p *Plugin) updateCounters() {
	if p.bufferCurSize < p.bufferMaxSize {
		p.bufferCurSize++
	}

	if p.bufferIndex == p.bufferMaxSize-1 {
		p.bufferIndex = 0
	} else {
		p.bufferIndex++
	}
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

	//logger := log.New()
	p.logger.Println("Statistics Processor started")

	var metrics []plugin.MetricType

	if config != nil {
		if config["SlidingWindowLength"].(ctypes.ConfigValueInt).Value > 0 {
			p.bufferMaxSize = config["SlidingWindowLength"].(ctypes.ConfigValueInt).Value
		} else {
			p.bufferMaxSize = 100
		}
	} else {
		p.bufferMaxSize = 100
	}

	//Decodes the content into PluginMetricType
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		p.logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for _, metric := range metrics {
		p.logger.Printf("Data received %v", metric.Data())
		p.insertInToBuffer(metric.Data(), metric.Namespace().Strings())
		p.updateCounters()

		//p.updateCounters()

		//	for _, metric := range metrics {
		var err error
		if p.bufferCurSize < p.bufferMaxSize {
			metric.Data_, err = p.calculateStats(p.buffer[concatNameSpace(metric.Namespace().Strings())][0:p.bufferCurSize])
			if err != nil {
				return "", nil, err
			}
		} else {
			metric.Data_, err = p.calculateStats(p.buffer[concatNameSpace(metric.Namespace().Strings())])
			if err != nil {
				return "", nil, err
			}
		}

		p.logger.Printf("Statistics %v", metric.Data())
		//}
		//	p.logger.Printf("Statistics %v", metrics)
	}

	gob.Register(map[string]float64{})
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(metrics); err != nil {
		return "", nil, err
	}

	return contentType, buf.Bytes(), nil
}
