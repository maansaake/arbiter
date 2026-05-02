package report

import (
	"context"
	"time"

	"tres-bon.se/arbiter/pkg/module"
)

type (
	Reporter interface {
		Start(context.Context)
		Op(module, op string, result *module.Result, err error)
		Finalise() error
	}

	TriggerReport[T ~int | ~uint | ~float64 | ~string] struct {
		Timestamp time.Time `yaml:"timestamp"`
		Type      string    `yaml:"type"`
		Value     T         `yaml:"value"`
	}
)
