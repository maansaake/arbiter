package module

import (
	"errors"
	"testing"
)

func TestMock(t *testing.T) {
	mod := NewMock()
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
		module        func() Module
		expectedError error
	}{
		{
			name: "reserved module name",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "arbiter"
				return mock
			},
			expectedError: ErrReservedPrefix,
		},
		{
			name: "invalid module name",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "invalid*"
				return mock
			},
			expectedError: ErrInvalidName,
		},
		{
			name: "valid module",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "valid-name"
				return mock
			},
			expectedError: nil,
		},
		{
			name: "invalid op name",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "valid"
				mock.SetOps = Ops{
					&Op{
						Name: "*invalid",
					},
				}
				return mock
			},
			expectedError: ErrInvalidName,
		},
		{
			name: "empty op name",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "valid"
				mock.SetOps = Ops{
					&Op{
						Name: "",
					},
				}
				return mock
			},
			expectedError: ErrInvalidName,
		},
		{
			name: "valid op",
			module: func() Module {
				mock := NewMock()
				mock.SetName = "valid"
				mock.SetOps = Ops{
					&Op{
						Name: "validname",
					},
				}
				return mock
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(Modules{test.module()})

			if test.expectedError != nil && !errors.Is(err, test.expectedError) {
				t.Fatalf("expected error %v, but got %v", test.expectedError, err)
			}
		})
	}
}
