package module

import (
	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

type MockModule struct {
	SetName string
	SetDesc string
	SetArgs arg.Args
	SetOps  op.Ops
}

var _ Module = &MockModule{}

func NewMock() *MockModule {
	return &MockModule{}
}

func (m *MockModule) Args() arg.Args {
	return m.SetArgs
}

func (m *MockModule) Desc() string {
	return m.SetDesc
}

func (m *MockModule) Name() string {
	return m.SetName
}

func (m *MockModule) Ops() op.Ops {
	return m.SetOps
}

func (m *MockModule) Run() error {
	return nil
}

func (m *MockModule) Stop() error {
	return nil
}
