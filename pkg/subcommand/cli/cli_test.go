package cli

import (
	"errors"
	"flag"
	"os"
	"testing"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
	testmodule "tres-bon.se/arbiter/pkg/module/test"
)

func intValidator(v int) bool {
	return v == 12
}

func TestHandleModule(t *testing.T) {
	count := &arg.Arg[int]{Name: "count", Required: true, Value: new(int)}
	master := &arg.Arg[bool]{Name: "master", Value: new(bool)}

	do := &op.Op{Name: "do"}
	more := &op.Op{Name: "more"}

	mod := &testmodule.TestModule{
		SetName: "mod",
		SetArgs: arg.Args{count, master},
		SetOps:  op.Ops{do, more},
	}
	os.Args = []string{"subcommand", "-mod.count=12", "-mod.do.rate=100", "-mod.more.disable"}

	err := Parse(0, module.Modules{mod})
	if err != nil {
		t.Fatal("should not have been an error")
	}
	if *count.Value != 12 {
		t.Fatal("should have been 12")
	}
	if *master.Value {
		t.Fatal("should have been false")
	}
	if do.Disabled {
		t.Fatal("do should be enabled")
	}
	if do.Rate != 100 {
		t.Fatal("do rate should have been 100")
	}
	if !more.Disabled {
		t.Fatal("more should be disabled")
	}
	if more.Rate != 0 {
		t.Fatal("more rate should have been 0")
	}
}

func TestRegisterRequired(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	register("ns", &arg.Arg[int]{
		Name:     "int",
		Value:    new(int),
		Required: true,
	})
	register("ns", &arg.Arg[float64]{
		Name:     "float",
		Value:    new(float64),
		Required: true,
	})
	register("ns", &arg.Arg[string]{
		Name:     "string",
		Value:    new(string),
		Required: true,
	})
	register("ns", &arg.Arg[bool]{
		Name:  "bool",
		Value: new(bool),
	})
	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}
	if len(required) != 3 {
		t.Fatal("should have had 3 required arguments")
	}
}

func TestRequiredPresent(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ExitOnError)
	required = []string{}

	register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})
	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	err := parse([]string{"-prefix.count=12"})
	if err != nil {
		t.Fatal("parsing failed:", err)
	}

	if len(required) != 0 {
		t.Fatal("number of required should have been cleared after parsing")
	}
}

func TestRequiredMissing(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ExitOnError)
	required = []string{}

	register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})
	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	err := parse([]string{})
	if err == nil {
		t.Fatal("parsing should have failed")
	}

	if !errors.Is(err, ErrParseError) {
		t.Fatal("expected a ErrParseError")
	}
}

func TestRequiredBoolean(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("should have panicked")
		}
	}()

	register("prefix", &arg.Arg[bool]{
		Name:     "master",
		Value:    new(bool),
		Required: true,
	})
}

func TestFlagParseFailure(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ContinueOnError)
	required = []string{}

	register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})

	err := parse([]string{"-doesnotexist=12"})
	if err == nil {
		t.Fatal("parsing should have failed")
	}
}

func TestParseInt(t *testing.T) {
	val := 12
	i := &arg.Arg[int]{
		Name:     "int",
		Value:    &val,
		Required: true,
	}

	err := intHandler("prefix", i)("13")
	if err != nil {
		t.Errorf("failed to handle int: %v", err)
	}

	if *i.Value != 13 {
		t.Fatal("value should have been 13")
	}
}

func TestParseFloat(t *testing.T) {
	val := 12.12
	fl := &arg.Arg[float64]{
		Name:  "float",
		Value: &val,
	}

	err := floatHandler("prefix", fl)("12.13")
	if err != nil {
		t.Errorf("failed to handle float: %v", err)
	}

	if *fl.Value != 12.13 {
		t.Fatal("value should have been 12.13")
	}
}

func TestParseString(t *testing.T) {
	val := "string"
	str := &arg.Arg[string]{
		Name:  "string",
		Value: &val,
	}

	err := stringHandler("prefix", str)("stringg")
	if err != nil {
		t.Errorf("failed to handle string: %v", err)
	}

	if *str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestParseIntFailure(t *testing.T) {
	err := intHandler("prefix", &arg.Arg[int]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseUIntFailure(t *testing.T) {
	err := uintHandler("prefix", &arg.Arg[uint]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseFloatFailure(t *testing.T) {
	err := floatHandler("prefix", &arg.Arg[float64]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestValidationError(t *testing.T) {
	b := &arg.Arg[int]{
		Name:  "bool",
		Value: new(int),
		Valid: func(val int) bool {
			return val != 12
		},
	}

	err := intHandler("prefix", b)("12")

	if err == nil {
		t.Fatal("should have forced a validation error")
	}
}

func TestValidationOk(t *testing.T) {
	iv := 13
	i := &arg.Arg[int]{
		Name:  "int",
		Value: &iv,
		Valid: intValidator,
	}

	err := intHandler("prefix", i)("12")
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if *i.Value != 12 {
		t.Fatal("should have been 12")
	}
}

func TestPanicUnrecognizedType(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	defer func() {
		_ = recover()
	}()
	register("ns", &arg.Arg[float32]{})
	t.Fatal("no panic? :(")
}

func TestPanicRegisterNilPointer(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	defer func() {
		_ = recover()
	}()
	register("ns", &arg.Arg[float64]{})
	t.Fatal("no panic? :(")
}

func TestParse(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	i := &arg.Arg[int]{
		Name:  "intt",
		Desc:  "desc",
		Value: new(int),
	}
	s := &arg.Arg[string]{
		Name:  "stringg",
		Desc:  "desc",
		Value: new(string),
	}
	register("ns", i)
	register("ns", s)
	op := &op.Op{
		Name: "op",
		Desc: "desc",
	}
	registerOp("ns", op)

	err := parse([]string{"-ns.stringg", "strvalue", "-ns.intt", "12", "-ns.op.disable", "-ns.op.rate", "12"})
	if err != nil {
		t.Fatal(err)
	}

	if *i.Value != 12 {
		t.Fatal("integer should have been 12")
	}

	if *s.Value != "strvalue" {
		t.Fatal("integer should have been 12")
	}

	if !op.Disabled {
		t.Fatal("should have been disabled")
	}

	if op.Rate != 12 {
		t.Fatal("rate should have been 12")
	}
}
