package zabbix

import (
	"fmt"
	"time"
)

type Config interface {
	ZabbixHost() string
	ZabbixSource() string
	ZabbixTimeout() time.Duration
}

type UpdatableConfig interface {
	Config
	WhenUpdated(f func() error)
}

func Empty(config Config) bool {
	return config.ZabbixHost() == "" && config.ZabbixSource() == ""
}

func PrintConfig(config Config) string {
	return fmt.Sprintf("zabbix[host='%s',timeout=%s,source='%s']", config.ZabbixHost(),
		config.ZabbixTimeout(), config.ZabbixSource())
}

func configValid(config Config) error {
	if Empty(config) {
		return nil
	}
	if config.ZabbixHost() == "" {
		return fmt.Errorf("%w: zabbix host is empty", ErrBadValue)
	}
	if timeout := config.ZabbixTimeout(); timeout == 0 {
		return fmt.Errorf("%w: empty zabbix timeout", ErrBadValue)
	}
	if config.ZabbixSource() == "" {
		return fmt.Errorf("%w: empty zabbix source", ErrBadValue)
	}
	return nil
}
