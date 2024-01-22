package zabbix

import "errors"

var (
	ErrBadValue     = errors.New("bad value")
	ErrBadResponse  = errors.New("bad response")
	ErrUnsupported  = errors.New("unsupported protocol")
	ErrUnsuccessful = errors.New("unsuccessful request")
)
