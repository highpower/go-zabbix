package zabbix

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type Sender struct {
	impl atomic.Value
}

type status struct {
	Response string `json:"response"`
	Info     string `json:"info"`
}

type senderImpl struct {
	host    string
	timeout time.Duration
}

//goland:noinspection SpellCheckingInspection

var zabbixHeader = []byte("ZBXD")

func (s *Sender) Send(metrics *Metrics) error {
	impl := s.impl.Load().(*senderImpl)
	st, err := send(impl.host, impl.timeout, metrics)
	if err != nil {
		return err
	}
	if err := statusValid(&st); err != nil {
		return err
	}
	return nil
}

func (s *Sender) setup(config Config) error {
	s.impl.Store(&senderImpl{host: config.ZabbixHost(), timeout: config.ZabbixTimeout()})
	return nil
}

func (s *status) String() string {
	return fmt.Sprintf("status[response='%s',info='%s']", s.Response, s.Info)
}

func New(config Config) (Sender, error) {
	result := Sender{}
	if err := result.setup(config); err != nil {
		return Sender{}, err
	}
	return result, nil
}

func send(host string, timeout time.Duration, metrics *Metrics) (status, error) {
	req, err := makeRequest(metrics)
	if err != nil {
		return status{}, err
	}
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return status{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err = conn.Write(req); err != nil {
		return status{}, err
	}
	resp, err := io.ReadAll(conn)
	if err != nil {
		return status{}, err
	}
	return parseResponse(resp)
}

func makeRequest(metrics *Metrics) ([]byte, error) {
	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	buffer := bytes.Buffer{}
	if _, err := buffer.Write(zabbixHeader); err != nil {
		return nil, err
	}
	if err := buffer.WriteByte(byte(1)); err != nil {
		return nil, err
	}
	fields := []uint32{uint32(len(data)), 0}
	for _, f := range fields {
		if err := binary.Write(&buffer, binary.LittleEndian, f); err != nil {
			return nil, err
		}
	}
	if _, err := buffer.Write(data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseResponse(resp []byte) (status, error) {
	buffer, err := readResponseHeader(resp)
	if err != nil {
		return status{}, err
	}
	length := uint32(0)
	if err := binary.Read(buffer, binary.LittleEndian, &length); err != nil {
		return status{}, err
	}
	reserved := uint32(0)
	if err := binary.Read(buffer, binary.LittleEndian, &reserved); err != nil {
		return status{}, err
	}
	if buffer.Len() != int(length) {
		return status{}, fmt.Errorf("%w has %d length instead of %d", ErrBadResponse, buffer.Len(), length)
	}
	st := status{}
	if err := json.NewDecoder(buffer).Decode(&st); err != nil {
		return status{}, err
	}
	return st, nil
}

func readResponseHeader(resp []byte) (*bytes.Buffer, error) {
	if len(resp) == 0 {
		return nil, fmt.Errorf("%w: empty", ErrBadResponse)
	}
	buffer := bytes.NewBuffer(resp)
	header := make([]byte, len(zabbixHeader))
	nb, err := buffer.Read(header)
	switch {
	case err != nil:
		return nil, err
	case nb != len(header):
		return nil, ErrBadResponse
	}
	if !bytes.Equal(zabbixHeader, header) {
		return nil, fmt.Errorf("%w does not start with %s", ErrBadResponse, string(zabbixHeader))
	}
	flag, err := buffer.ReadByte()
	switch {
	case err != nil:
		return nil, err
	case flag != 1:
		return nil, ErrUnsupported
	}
	return buffer, nil
}

func statusValid(st *status) error {
	if !strings.EqualFold(st.Response, "success") {
		return fmt.Errorf("%w: response is %s", ErrUnsuccessful, st.Response)
	}
	return nil
}
