package metric

type Metric interface {
	Pull() ([]byte, error)
}
