package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/highpower/go-zabbix"
	"os"
	"time"
)

type config struct {
	Host    string
	Source  string
	Timeout time.Duration
}

var (
	errBadConfig = errors.New("incorrect config")
)

func (c *config) ZabbixHost() string {
	return c.Host
}

func (c *config) ZabbixTimeout() time.Duration {
	return c.Timeout
}

func (c *config) ZabbixSource() string {
	return c.Source
}

func (c *config) valid() error {
	if c.Host == "" {
		return fmt.Errorf("%w: empty zabbix host", errBadConfig)
	}
	return nil
}

func sendToZabbix(c *config, metric string) error {
	sender, err := zabbix.NewSender(c)
	if err != nil {
		return err
	}
	metrics := zabbix.NewMetrics(c.Source)
	metrics.Add(metric, 1)
	return sender.Send(&metrics)
}

func halt(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}

func main() {

	var conf config
	var metric string

	flag.StringVar(&conf.Host, "host", "", "host to connect to zabbix server")
	flag.StringVar(&conf.Source, "source", "", "source if differs from host name")
	flag.DurationVar(&conf.Timeout, "timeout", 150*time.Millisecond, "timeout to connect to zabbix server")
	flag.StringVar(&metric, "metric", "", "metric name to send")

	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}
	if err := conf.valid(); err != nil {
		halt(err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", zabbix.PrintConfig(&conf))
	if err := sendToZabbix(&conf, metric); err != nil {
		halt(err)
	}
}
