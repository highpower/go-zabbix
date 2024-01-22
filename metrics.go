package zabbix

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Metrics struct {
	source string
	when   time.Time
	values map[string]any
}

type metricBuffer struct {
	Host  string      `json:"host"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Clock int64       `json:"clock"`
}

type packetBuffer struct {
	Request string         `json:"request"`
	Data    []metricBuffer `json:"data"`
	Clock   int64          `json:"clock"`
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	clock := m.when.Unix()
	packet := packetBuffer{Request: "sender data", Clock: clock}
	for k, v := range m.values {
		metric := metricBuffer{Host: m.source, Key: k, Value: v, Clock: clock}
		packet.Data = append(packet.Data, metric)
	}
	return json.Marshal(&packet)
}

func (m *Metrics) String() string {
	return fmt.Sprintf("Metrics[source='%s',when='%s',values=[%s]", m.source, m.when.Format(time.RFC822),
		printValueMap(m.values))
}

func (m *Metrics) Add(key string, value any) {
	m.values[key] = value
}

func NewMetrics(source string) Metrics {
	return Metrics{source: source, values: make(map[string]interface{}), when: time.Now()}
}

func printValueMap(m map[string]any) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, printValue(k, v))
	}
	return fmt.Sprintf("Values[%s]", strings.Join(parts, ","))
}

func printValue(key string, value any) string {
	var result string
	if m, ok := value.(encoding.TextMarshaler); ok {
		result = textValue(m)
	} else {
		result = fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("value[key='%s',value='%s']", key, result)
}

func textValue(value encoding.TextMarshaler) string {
	b, err := value.MarshalText()
	if err != nil {
		panic(fmt.Errorf("%w: error (%s) occurred in text marshall", ErrBadValue, err.Error()))
	}
	return string(b)
}
