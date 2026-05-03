package modulemock

import "github.com/maansaake/arbiter/pkg/module"

type Module struct {
	SetName string
	SetDesc string
	SetArgs module.Args
	SetOps  module.Ops
}

var _ module.Module = &Module{}

func NewMock() *Module {
	return &Module{}
}

func (m *Module) Args() module.Args {
	return m.SetArgs
}

func (m *Module) Desc() string {
	return m.SetDesc
}

func (m *Module) Name() string {
	return m.SetName
}

func (m *Module) Ops() module.Ops {
	return m.SetOps
}

func (m *Module) Run() error {
	return nil
}

func (m *Module) Stop() error {
	return nil
}
