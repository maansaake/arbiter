package report

type YAMLReporter struct{}

func NewYAML() Reporter {
	return &YAMLReporter{}
}

func (r *YAMLReporter) Finalise() {}
