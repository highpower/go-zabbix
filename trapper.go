package zabbix

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Var interface {
	String() string
}

type Func func(key string, value Var)

type VarMap interface {
	Name() string
	ForEach(f Func)
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

type Trapper struct {
	prefix string
	log    Logger
	impl   atomic.Value
}

type trapperImpl struct {
	host    string
	timeout time.Duration
	source  string
}

func (t *Trapper) SendEvery(period time.Duration, vars ...VarMap) (Stopper, error) {
	ticker := time.NewTicker(period)
	go t.runSend(ticker.C, vars)
	return ticker, nil
}

func (t *Trapper) runSend(c <-chan time.Time, vars []VarMap) {
	for range c {
		impl := t.impl.Load().(*trapperImpl)
		Debugf(t.log, "trapper.runSend: sending metrics to %s", impl.host)
		metrics := NewMetrics(impl.source)
		for _, vm := range vars {
			vm.ForEach(func(name string, value Var) {
				metrics.Add(fmt.Sprintf("%s.%s.%s", t.prefix, vm.Name(), name), value)
			})
		}
		if _, err := send(impl.host, impl.timeout, &metrics); err != nil {
			Errorf(t.log, "trapper.runSend: %s", err.Error())
		}
	}
}

func (t *Trapper) setup(config Config) error {
	if err := configValid(config); err != nil {
		return err
	}
	t.impl.Store(&trapperImpl{host: config.ZabbixHost(), timeout: config.ZabbixTimeout(),
		source: config.ZabbixSource()})
	return nil
}

func NewTrapper(config UpdatableConfig, log Logger, prefix string) (Trapper, error) {
	result := Trapper{prefix: prefix, log: log}
	if err := result.setup(config); err != nil {
		return Trapper{}, err
	}
	config.WhenUpdated(func() error { return result.setup(config) })
	return result, nil
}

func Stop(stopper Stopper) {
	if stopper != nil {
		stopper.Stop()
	}
}

func Errorf(log Logger, format string, v ...any) {
	if log != nil {
		log.Errorf(format, v...)
	}
}

func Debugf(log Logger, format string, v ...any) {
	if log != nil {
		log.Debugf(format, v...)
	}
}
