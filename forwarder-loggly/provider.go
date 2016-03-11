package main

import (
	"net/http"

	"github.com/glerchundi/journald-forwarder/core"
)

type LogglyProviderConfig struct {
	Token string
}

func NewLogglyProvider(config LogglyProviderConfig) (*LogglyProvider, error) {
	return &LogglyProvider{
		client: &http.Client{},
		endpoint: "https://logs-01.loggly.com/bulk/" + config.Token,
	}, nil
}

type LogglyProvider struct {
	client   *http.Client
	endpoint string
}

func (lp *LogglyProvider) GetBulkSize() int {
	return 1
}

func (lp *LogglyProvider) Publish(iterator core.JournalEntryIterator) (int, error) {

	/*
	// Propagate!
	req, err := http.NewRequest("POST", lp.endpoint, bytes.NewBuffer(body))
	if err != nil {
		debug("error: %v", err)
		return err
	}

	req.Header.Add("User-Agent", "go-loggly (version: "+Version+")")
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Content-Length", string(len(body)))

	tags := c.tagsList()
	if tags != "" {
		req.Header.Add("X-Loggly-Tag", tags)
	}

	res, err := client.Do(req)
	if err != nil {
		debug("error: %v", err)
		return err
	}

	defer res.Body.Close()

	debug("%d response", res.StatusCode)
	if res.StatusCode >= 400 {
		resp, _ := ioutil.ReadAll(res.Body)
		debug("error: %s", string(resp))
	}

	return err
*/

	/*
	index := 0
	for iterator.Next() {
		i, e := iterator.Value()
		os.Stdout.WriteString(e.Cursor)
		os.Stdout.Write([]byte{'\n'})
		index = i
	}

	time.Sleep(1 * time.Second)
	return index, nil
	*/

	return 0, nil
}