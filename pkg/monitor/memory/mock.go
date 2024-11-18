package memory

type MemoryMock struct {
	Memory

	onRead func()
	vms    uint
	rss    uint
	err    error
}

func NewMemoryMock(vms, rss uint, err error) *MemoryMock {
	return &MemoryMock{
		vms: vms,
		rss: rss,
		err: err,
	}
}

func (m *MemoryMock) VMS() (uint, error) {
	if m.onRead != nil {
		m.onRead()
	}
	return m.vms, m.err
}

func (m *MemoryMock) RSS() (uint, error) {
	if m.onRead != nil {
		m.onRead()
	}
	return m.rss, m.err
}

func (m *MemoryMock) SetVMS(newVal uint) {
	m.vms = newVal
}

func (m *MemoryMock) SetRSS(newVal uint) {
	m.rss = newVal
}

func (m *MemoryMock) SetErr(newErr error) {
	m.err = newErr
}

func (m *MemoryMock) SetOnRead(onRead func()) {
	m.onRead = onRead
}
