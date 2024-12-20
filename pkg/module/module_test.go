package module

import (
	"errors"
	"testing"

	"tres-bon.se/arbiter/pkg/module/op"
)

func TestValidateModules(t *testing.T) {
	tests := []struct {
		name          string
		module        func() Module
		expectedError error
	}{
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
				mock.SetOps = op.Ops{
					&op.Op{
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
				mock.SetOps = op.Ops{
					&op.Op{
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
				mock.SetOps = op.Ops{
					&op.Op{
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