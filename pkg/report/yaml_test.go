package report

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestMarshalReport(t *testing.T) {
	report := &report{
		Start:    time.Now(),
		End:      time.Now(),
		Duration: 5 * time.Second,
	}

	bs, err := yaml.Marshal(report)
	t.Log(err)
	t.Log(string(bs))
}
