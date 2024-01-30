package zabbix

import (
	"fmt"
	"time"
)

type SenderConfig interface {
	ZabbixHost() string
	ZabbixTimeout() time.Duration
}

type Config interface {
	SenderConfig
	ZabbixSource() string
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

func PrintSenderConfig(config SenderConfig) string {
	return fmt.Sprintf("zabbix[host='%s',timeout=%s]", config.ZabbixHost(), config.ZabbixTimeout())
}

func senderConfigValid(config SenderConfig) error {
	if config.ZabbixHost() == "" {
		return fmt.Errorf("%w: zabbix host is empty", ErrBadValue)
	}
	if timeout := config.ZabbixTimeout(); timeout == 0 {
		return fmt.Errorf("%w: empty zabbix timeout", ErrBadValue)
	}
	return nil
}

func configValid(config Config) error {
	if Empty(config) {
		return nil
	}
	if err := senderConfigValid(config); err != nil {
		return err
	}
	if config.ZabbixSource() == "" {
		return fmt.Errorf("%w: empty zabbix source", ErrBadValue)
	}
	return nil
}
