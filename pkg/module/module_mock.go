package module

type MockModule struct {
	SetName string
	SetDesc string
	SetArgs Args
	SetOps  Ops
}

var _ Module = &MockModule{}

func NewMock() *MockModule {
	return &MockModule{}
}

func (m *MockModule) Args() Args {
	return m.SetArgs
}

func (m *MockModule) Desc() string {
	return m.SetDesc
}

func (m *MockModule) Name() string {
	return m.SetName
}

func (m *MockModule) Ops() Ops {
	return m.SetOps
}

func (m *MockModule) Run() error {
	return nil
}

func (m *MockModule) Stop() error {
	return nil
}
