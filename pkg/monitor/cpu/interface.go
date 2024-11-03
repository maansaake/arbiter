package cpu

type CPU interface {
	Read() (float64, error)
}
