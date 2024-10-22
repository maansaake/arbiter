package zerologr

import "testing"

func init() {
	VFieldName = "v"
}

func TestGlobal(t *testing.T) {
	SetLogger(New(&Opts{Console: true, Caller: true}).WithName("global"))
	Info("global logger")
	V(1).Info("not logged")
	V(0).Info("logged")
	Info("post")
}

func TestLocal(t *testing.T) {
	logger := New(&Opts{Console: true, Caller: true}).WithName("local")
	logger.Info("local logger")
	logger.V(1).Info("not logged")
	logger.V(0).Info("logged")
	logger.WithName("newname").Info("hi")
	logger.WithValues("val", 12, "notherval", 12.12).Info("values")
	logger.Info("post")
}
