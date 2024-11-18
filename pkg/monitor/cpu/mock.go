package cpu

type CPUMock struct {
	CPU

	onRead func()
	val    float64
	err    error
}

// Read implements CPU.
func (m *CPUMock) Read() (float64, error) {
	if m.onRead != nil {
		m.onRead()
	}
	return m.val, m.err
}

func NewCPUMonitorMock(val float64, err error) *CPUMock {
	return &CPUMock{val: val, err: err}
}

func (m *CPUMock) SetVal(newVal float64) {
	m.val = newVal
}

func (m *CPUMock) SetErr(newErr error) {
	m.err = newErr
}

func (m *CPUMock) SetOnRead(onRead func()) {
	m.onRead = onRead
}
