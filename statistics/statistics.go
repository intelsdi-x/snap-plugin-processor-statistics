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
	"strings"

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

// New() returns a new instance of this
func New() *Plugin {
	buffer := make(map[string][]interface{})
	p := &Plugin{buffer: buffer,
		bufferMaxSize: 100,
		bufferCurSize: 0,
		bufferIndex:   0}
	return p
}

// calculateStats calaculates the descriptive statistics for buff
func (p *Plugin) calculateStats(buff interface{}, logger *log.Logger) (map[string][]float64, error) {
	result := make(map[string][]float64)

	var buffer []float64

	logger.Printf("Buff %v", buff)

	var valRange []float64 //stores range values
	for _, val := range buff.([]interface{}) {
		switch v := val.(type) {
		default:
			logger.Printf("Unknown data received: Type %T", v)
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

	logger.Printf("Buffer %v", buffer)

	result["Count"] = []float64{float64(len(buffer))}

	val, err := stats.Mean(buffer)
	if err != nil {
		return nil, err
	}

	result["mean"] = []float64{val}

	val, err = stats.Median(buffer)
	if err != nil {
		return nil, err
	}

	result["median"] = []float64{val}

	val, err = stats.StandardDeviation(buffer)
	if err != nil {
		return nil, err
	}

	result["stddev"] = []float64{val}

	val, err = stats.Variance(buffer)
	if err != nil {
		return nil, err
	}

	result["var"] = []float64{val}

	val, err = stats.Percentile(buffer, 95)
	if err != nil {
		return nil, err
	}

	result["95%-ile"] = []float64{val}

	val, err = stats.Percentile(buffer, 99)
	if err != nil {
		return nil, err
	}

	result["99%-ile"] = []float64{val}

	val, err = stats.Min(buffer)
	if err != nil {
		return nil, err
	}

	result["Minimum"] = []float64{val}
	valRange = append(valRange, val)

	val, err = stats.Max(buffer)
	if err != nil {
		return nil, err
	}

	result["Maximum"] = []float64{val}
	valRange = append(valRange, val)

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

	result["Kurtosis"] = p.Kurtosis(buffer, logger)
	result["Skewness"] = p.Skewness(buffer, logger)

	result["Sum"] = []float64{val}

	val, err = stats.Trimean(buffer)
	if err != nil {
		return nil, err
	}

	result["Trimean"] = []float64{val}
	return result, nil
}

//Calculates the population skewness from buffer
func (p *Plugin) Skewness(buffer []float64, logger *log.Logger) []float64 {
	if len(buffer) == 0 {
		logger.Printf("Buffer does not contain any data.")
		return []float64{}
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

	return []float64{math.Sqrt(float64(len(buffer))) * num / math.Pow(den, 3/2)}

}

//Calculates the population kurtosis from buffer
func (p *Plugin) Kurtosis(buffer []float64, logger *log.Logger) []float64 {
	if len(buffer) == 0 {
		logger.Printf("Buffer does not contain any data.")
		return []float64{}
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
	return []float64{float64(len(buffer)) * num / math.Pow(den, 2)}
}

// concatNameSpace combines an array of namespces into a single string
func concatNameSpace(namespace []string) string {
	completeNamespace := strings.Join(namespace, ", ")
	return completeNamespace
}

// insertInToBuffer adds a new value into this' buffer object
func (p *Plugin) insertInToBuffer(val interface{}, ns []string, logger *log.Logger) {

	if p.bufferCurSize == 0 {
		logger.Printf("In if statement with buffer max size %v", p.bufferMaxSize)
		var buff = make([]interface{}, p.bufferMaxSize)
		buff[0] = val
		p.buffer[concatNameSpace(ns)] = buff
	} else {
		logger.Printf("In else statement with index %v", p.bufferIndex)
		p.buffer[concatNameSpace(ns)][p.bufferIndex] = val
	}
}

// updateCounters updates the meta informaiton (current size and index) of this' buffer object
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

// GetConfigPolicy returns the config policy
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

// Process processes the data, inputs the data into this' buffer and calls the descriptive statistics method
func (p *Plugin) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	logger := log.New()
	logger.Println("Statistics Processor started")

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
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for _, metric := range metrics {
		logger.Printf("Data received %v", metric.Data())
		p.insertInToBuffer(metric.Data(), metric.Namespace().Strings(), logger)
		p.updateCounters()

		var err error
		if p.bufferCurSize < p.bufferMaxSize {
			metric.Data_, err = p.calculateStats(p.buffer[concatNameSpace(metric.Namespace().Strings())][0:p.bufferCurSize], logger)
			if err != nil {
				return "", nil, err
			}
		}
		metric.Data_, err = p.calculateStats(p.buffer[concatNameSpace(metric.Namespace().Strings())], logger)
		if err != nil {
			return "", nil, err
		}

		logger.Printf("Statistics %v", metric.Data())
	}

	gob.Register(map[string]float64{})
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(metrics); err != nil {
		return "", nil, err
	}

	return contentType, buf.Bytes(), nil
}
