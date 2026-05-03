package module_test

import (
	"errors"
	"testing"

	"github.com/maansaake/arbiter/pkg/module"
	modulemock "github.com/maansaake/arbiter/pkg/module/mock"
)

func TestMock(t *testing.T) {
	mod := modulemock.NewMock()
	_ = mod.Name()
	_ = mod.Desc()
	_ = mod.Ops()
	_ = mod.Args()
	if err := mod.Run(); err != nil {
		t.Fatal(err)
	}
	if err := mod.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateModules(t *testing.T) {
	tests := []struct {
		name          string
		module        func() module.Module
		expectedError error
	}{
		{
			name: "reserved module name",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "arbiter"
				return mod
			},
			expectedError: module.ErrReservedPrefix,
		},
		{
			name: "invalid module name",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "invalid*"
				return mod
			},
			expectedError: module.ErrInvalidName,
		},
		{
			name: "valid module",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "valid-name"
				return mod
			},
			expectedError: nil,
		},
		{
			name: "invalid op name",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "valid"
				mod.SetOps = module.Ops{
					&module.Op{
						Name: "*invalid",
					},
				}
				return mod
			},
			expectedError: module.ErrInvalidName,
		},
		{
			name: "empty op name",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "valid"
				mod.SetOps = module.Ops{
					&module.Op{
						Name: "",
					},
				}
				return mod
			},
			expectedError: module.ErrInvalidName,
		},
		{
			name: "valid op",
			module: func() module.Module {
				mod := modulemock.NewMock()
				mod.SetName = "valid"
				mod.SetOps = module.Ops{
					&module.Op{
						Name: "validname",
					},
				}
				return mod
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := module.Validate(module.Modules{test.module()})

			if test.expectedError != nil && !errors.Is(err, test.expectedError) {
				t.Fatalf("expected error %v, but got %v", test.expectedError, err)
			}
		})
	}
}
