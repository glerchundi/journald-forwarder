package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/glerchundi/journald-forwarder/core"
)

type LogglyProviderConfig struct {
	Token string
	Tags  string
}

func (*LogglyProviderConfig) Name() string {
	return "loggly"
}

func (*LogglyProviderConfig) BulkSize() int {
	return 1
}

func NewLogglyProviderConfig() *LogglyProviderConfig {
	return &LogglyProviderConfig{}
}

type LogglyProvider struct {
	client     *http.Client
	endpoint   string
	tags       string
	marshaller core.JournalEntryMarshaller
}

func NewLogglyProvider(config *LogglyProviderConfig) (*LogglyProvider, error) {
	if config.Token == "" {
		return nil, errors.New("token not provided")
	}

	return &LogglyProvider{
		client:     &http.Client{},
		endpoint:   "https://logs-01.loggly.com/bulk/" + config.Token,
		tags:       config.Tags,
		marshaller: core.JournalEntryMarshaller{},
	}, nil
}

func (lp *LogglyProvider) Publish(iterator core.JournalEntryIterator) (int, error) {
	if !iterator.Next() {
		return 0, nil
	}

	_, e := iterator.Value()
	body := lp.marshaller.MarshalOne(e)

	// propagate!
	req, err := http.NewRequest("POST", lp.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}

	req.Header.Add("User-Agent", "journald-forwarder (version: 0.1.0)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", string(len(body)))

	if lp.tags != "" {
		req.Header.Add("X-Loggly-Tag", lp.tags)
	}

	res, err := lp.client.Do(req)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		resp, _ := ioutil.ReadAll(res.Body)
		return -1, fmt.Errorf("failed to post to loggly: %v", resp)
	}

	return 1, nil
}