package reporter

type Reporter interface {
	Report() error
}

type YAMLReporter struct{}

func (yr *YAMLReporter) Report() error { return nil }
