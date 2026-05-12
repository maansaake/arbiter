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
	t.Run("reserved module name", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "arbiter"
		err := module.Validate(module.Modules{mod})
		if !errors.Is(err, module.ErrReservedPrefix) {
			t.Fatalf("expected error %v, but got %v", module.ErrReservedPrefix, err)
		}
	})

	t.Run("invalid module name", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "invalid*"
		err := module.Validate(module.Modules{mod})
		if !errors.Is(err, module.ErrInvalidName) {
			t.Fatalf("expected error %v, but got %v", module.ErrInvalidName, err)
		}
	})

	t.Run("valid module", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "valid-name"
		err := module.Validate(module.Modules{mod})
		if err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}
	})

	t.Run("invalid op name", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "valid"
		mod.SetOps = module.Ops{
			&module.Op{Name: "*invalid"},
		}
		err := module.Validate(module.Modules{mod})
		if !errors.Is(err, module.ErrInvalidName) {
			t.Fatalf("expected error %v, but got %v", module.ErrInvalidName, err)
		}
	})

	t.Run("empty op name", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "valid"
		mod.SetOps = module.Ops{
			&module.Op{Name: ""},
		}
		err := module.Validate(module.Modules{mod})
		if !errors.Is(err, module.ErrInvalidName) {
			t.Fatalf("expected error %v, but got %v", module.ErrInvalidName, err)
		}
	})

	t.Run("valid op", func(t *testing.T) {
		mod := modulemock.NewMock()
		mod.SetName = "valid"
		mod.SetOps = module.Ops{
			&module.Op{Name: "validname"},
		}
		err := module.Validate(module.Modules{mod})
		if err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}
	})
}
