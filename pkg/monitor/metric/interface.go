package metric

type Metric interface {
	Pull()
}
