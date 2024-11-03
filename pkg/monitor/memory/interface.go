package memory

type Memory interface {
	RSS() (uint, error)
	VMS() (uint, error)
}
