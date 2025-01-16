package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/maansaake/locksmith/pkg/client"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type locksmithModule struct {
	args module.Args
	ops  module.Ops

	c        client.Client
	lockTags chan string
}

var (
	_ module.Module = &locksmithModule{}

	address = &module.Arg[string]{
		Name:  "address",
		Desc:  "Address of the locksmith server.",
		Value: new(string),
	}
	port = &module.Arg[uint]{
		Name:  "port",
		Desc:  "Port of the locksmith server.",
		Value: new(uint),
	}

	logger = zerologr.New(&zerologr.Opts{Console: true, V: 10}).WithName("locksmith")
)

func newLocksmithModule() *locksmithModule {
	lm := &locksmithModule{
		args:     module.Args{address, port},
		lockTags: make(chan string, 1000),
	}
	lm.ops = module.Ops{
		&module.Op{
			Name: "lock",
			Desc: "Locks a random lock.",
			Rate: 60,
			Do:   lm.lock,
		},
		&module.Op{
			Name: "unlock",
			Desc: "Unlocks a random lock.",
			Rate: 60,
			Do:   lm.unlock,
		},
	}

	return lm
}

func (lm *locksmithModule) Name() string {
	return "locksmith"
}

func (lm *locksmithModule) Desc() string {
	return "This is a sample module with a few sample operations."
}

func (lm *locksmithModule) Args() module.Args {
	return lm.args
}

func (lm *locksmithModule) Ops() module.Ops {
	return lm.ops
}

func (lm *locksmithModule) Run() error {
	logger.Info("running locksmith module")

	lm.c = client.New(&client.Opts{
		Host: *address.Value,
		//nolint:gosec // just for show
		Port:       uint16(*port.Value),
		OnAcquired: lm.onAcquired,
	})

	return lm.c.Connect()
}

func (lm *locksmithModule) Stop() error {
	lm.c.Close()
	return nil
}

func (lm *locksmithModule) lock() (module.Result, error) {
	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		return module.Result{}, err
	}

	lockTag := base64.URLEncoding.EncodeToString(bs)
	logger.V(100).Info("acquiring lock", "lock", lockTag)

	return module.Result{}, lm.c.Acquire(lockTag)
}

func (lm *locksmithModule) unlock() (module.Result, error) {
	if len(lm.lockTags) == 0 {
		logger.Info("no locks to unlock")
		return module.Result{}, nil
	}

	lockTag := <-lm.lockTags
	logger.V(100).Info("unlocking", "lock", lockTag)

	return module.Result{}, lm.c.Release(lockTag)
}

func (lm *locksmithModule) onAcquired(lockTag string) {
	logger.V(100).Info("lock acquired", "lock", lockTag, "acquired", len(lm.lockTags))

	lm.lockTags <- lockTag
}
