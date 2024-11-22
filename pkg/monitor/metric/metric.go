package metric

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type metric struct {
	url string
}

func NewMetricMonitor(url string) Metric {
	if !(strings.Contains(url, "http://") || strings.Contains(url, "https://")) {
		url = fmt.Sprintf("http://%s", url)
	}
	return &metric{
		url: url,
	}
}

func (m *metric) Pull() ([]byte, error) {
	if resp, err := http.Get(m.url); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		if bs, err := io.ReadAll(resp.Body); err != nil {
			return nil, err
		} else {
			return bs, nil
		}
	}
}
