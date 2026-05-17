// Package collection provides a reporter that fans out to multiple reporters.
package collection

import (
	"context"
	"errors"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
)

// reporter dispatches every reporter interface call to each of its child
// reporters in registration order.
type reporter struct {
	reporters []report.Reporter
}

var _ report.Reporter = &reporter{}

// New returns a Reporter that delegates to each of the provided reporters.
func New(reporters ...report.Reporter) report.Reporter {
	return &reporter{reporters: reporters}
}

// Start implements report.Reporter.
func (r *reporter) Start(ctx context.Context) {
	for _, rep := range r.reporters {
		rep.Start(ctx)
	}
}

// ReportError implements report.Reporter.
func (r *reporter) ReportError(err error) {
	for _, rep := range r.reporters {
		rep.ReportError(err)
	}
}

// ReportOp implements report.Reporter.
func (r *reporter) ReportOp(mod, op string, res *module.Result, err error) {
	for _, rep := range r.reporters {
		rep.ReportOp(mod, op, res, err)
	}
}

// Finalise implements report.Reporter. Reporters are finalised in registration
// order; all errors are joined and returned.
func (r *reporter) Finalise() error {
	var errs []error
	for _, rep := range r.reporters {
		if err := rep.Finalise(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
