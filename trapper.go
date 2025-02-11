package zabbix

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

type Var interface {
	String() string
}

type VarMap[C Var] interface {
	Name() string
	ForEach(f func(key string, value C))
}

type Logger interface {
	Fatalf(format string, v ...any)
	Errorf(format string, v ...any)
	Infof(format string, v ...any)
	Debugf(format string, v ...any)
}

type Stopper interface {
	Stop()
}

type Trapper[C fmt.Stringer] struct {
	prefix string
	log    Logger
	impl   atomic.Value
}

type trapperImpl struct {
	host    string
	timeout time.Duration
	source  string
}

func (t *Trapper[C]) SendEvery(period time.Duration, vars []VarMap[C]) (Stopper, error) {
	ticker := time.NewTicker(period)
	go t.runSend(ticker.C, vars)
	return ticker, nil
}

func (t *Trapper[C]) SendValuesEvery(period time.Duration, vars ...VarMap[C]) (Stopper, error) {
	return t.SendEvery(period, vars)
}

func (t *Trapper[C]) runSend(c <-chan time.Time, vars []VarMap[C]) {
	infof(t.log, "trapper.runSend: starting")
	defer infof(t.log, "trapper.runSend: stopping")
	for range c {
		impl := t.impl.Load().(*trapperImpl)
		debugf(t.log, "trapper.runSend: sending metrics to %s", impl.host)
		metrics := NewMetrics(impl.source)
		for _, vm := range vars {
			vm.ForEach(func(name string, value C) {
				key := fmt.Sprintf("%s.%s.%s", t.prefix, vm.Name(), name)
				debugf(t.log, "trapper.runSend: adding %s=%v", key, value)
				metrics.Add(key, value)
			})
		}
		if _, err := send(impl.host, impl.timeout, &metrics); err != nil {
			errorf(t.log, "trapper.runSend: %s", err.Error())
		}
	}
}

func (t *Trapper[C]) setup(config Config) error {
	if err := configValid(config); err != nil {
		return err
	}
	t.impl.Store(&trapperImpl{host: config.ZabbixHost(), timeout: config.ZabbixTimeout(),
		source: config.ZabbixSource()})
	return nil
}

func NewTrapper[C fmt.Stringer](config UpdatableConfig, log Logger, prefix string) (Trapper[C], error) {
	result := Trapper[C]{prefix: prefix, log: log}
	if err := result.setup(config); err != nil {
		return Trapper[C]{}, err
	}
	config.WhenUpdated(func() error { return result.setup(config) })
	return result, nil
}

func Stop(stopper Stopper) {
	if !isNil(stopper) {
		stopper.Stop()
	}
}

func infof(log Logger, format string, v ...any) {
	if !isNil(log) {
		log.Infof(format, v...)
	}
}

func errorf(log Logger, format string, v ...any) {
	if !isNil(log) {
		log.Errorf(format, v...)
	}
}

func debugf(log Logger, format string, v ...any) {
	if !isNil(log) {
		log.Debugf(format, v...)
	}
}

func isNil(v any) bool {
	return v == nil || (reflect.TypeOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
