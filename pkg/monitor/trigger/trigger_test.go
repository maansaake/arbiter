package trigger

import "testing"

func TestResultString(t *testing.T) {
	tests := []Result{NOTHING, RAISE, CLEAR}
	for _, res := range tests {
		if res.String() == UNRECOGNIZED_RESULT {
			t.Fatal("seems you missed to add this result to the String() method", res)
		}
	}
}

func TestRaiseClearInt(t *testing.T) {
	trigger := New(&TriggerOpts[int]{
		TriggerOn: ABOVE,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})

	if trigger.Update(11) != RAISE {
		t.Fatal("should be raise")
	}
	if trigger.Update(11) != NOTHING {
		t.Fatal("should be nothing")
	}
	if trigger.Update(10) != CLEAR {
		t.Fatal("should be clear")
	}
	if trigger.Update(9) != NOTHING {
		t.Fatal("should be nothing")
	}
	if trigger.Update(10000) != RAISE {
		t.Fatal("should be raise")
	}
	if trigger.Update(400) != NOTHING {
		t.Fatal("should be nothing")
	}
}

func TestRaiseClearString(t *testing.T) {
	trigger := New(&TriggerOpts[string]{
		TriggerOn: EQUAL,
		Raise:     "raising",
		SendClear: true,
		Clear:     "clearing",
	})

	if trigger.Update("raising") != RAISE {
		t.Fatal("should be raise")
	}
	if trigger.Update("raising") != RAISE {
		t.Fatal("should be raise")
	}
	if trigger.Update("clearing") != CLEAR {
		t.Fatal("should be clear")
	}
}

func TestStringAbove(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("a panic should have been raised")
		}
	}()
	New(&TriggerOpts[string]{
		TriggerOn: ABOVE,
		Raise:     "raising",
		SendClear: true,
		Clear:     "clearing",
	})
}

func TestAboveRaiseLowerThanClear(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("a panic should have been raised")
		}
	}()
	New(&TriggerOpts[int]{
		TriggerOn: ABOVE,
		Raise:     8,
		SendClear: true,
		Clear:     10,
	})
}

func TestAboveOrEqualRaiseEqualToClear(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("a panic should have been raised")
		}
	}()
	New(&TriggerOpts[int]{
		TriggerOn: ABOVE_OR_EQUAL,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})
}

func TestAboveOrEqualRaiseClear(t *testing.T) {
	trigger := New(&TriggerOpts[int]{
		TriggerOn: ABOVE_OR_EQUAL,
		Raise:     11,
		SendClear: true,
		Clear:     10,
	})

	if trigger.Update(11) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trigger.Update(10) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
	if trigger.Update(10) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(0) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(11) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trigger.Update(110000) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(1) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
	if trigger.Update(1) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
}

func TestBelowRaiseLowerThanClear(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("should have panicked")
		}
	}()

	New(&TriggerOpts[int]{
		TriggerOn: BELOW,
		Raise:     10,
		SendClear: true,
		Clear:     9,
	})
}

func TestBelowRaiseClear(t *testing.T) {
	trigger := New(&TriggerOpts[int]{
		TriggerOn: BELOW,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})

	if trigger.Update(10) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(1) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trigger.Update(5) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(10) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
	if trigger.Update(1) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trigger.Update(10000) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
	if trigger.Update(10) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
}

func TestBelowOrEqualClearEqualToRaise(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("should have panicked")
		}
	}()

	New(&TriggerOpts[int]{
		TriggerOn: BELOW_OR_EQUAL,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})
}

func TestBelowOrEqualRaiseClear(t *testing.T) {
	trigger := New(&TriggerOpts[int]{
		TriggerOn: BELOW_OR_EQUAL,
		Raise:     100,
		SendClear: true,
		Clear:     1000,
	})

	if trigger.Update(500) != NOTHING {
		t.Fatal("should have been NOTHING")
	}

	if trigger.Update(1100) != NOTHING {
		t.Fatal("should have been NOTHING")
	}

	if trigger.Update(99) != RAISE {
		t.Fatal("should have been RAISE")
	}

	if trigger.Update(999) != NOTHING {
		t.Fatal("should have been NOTHING")
	}

	if trigger.Update(1001) != CLEAR {
		t.Fatal("should have been CLEAR")
	}

	if trigger.Update(1000) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
}

func TestEqualRaiseClear(t *testing.T) {
	trigger := New(&TriggerOpts[int]{
		TriggerOn: EQUAL,
		Raise:     100,
		SendClear: true,
		Clear:     1000,
	})

	if trigger.Update(100) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trigger.Update(100) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(99) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(101) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(999) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(1000) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
	if trigger.Update(1000) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trigger.Update(1001) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
}
