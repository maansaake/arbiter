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
	trigger := New(&Opts[int]{
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
	trigger := triggerImpl[string]{
		triggerOn: EQUAL,
		raise:     "raising",
		sendClear: true,
		clear:     "clearing",
	}

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
	err := Validate(&Opts[string]{
		TriggerOn: ABOVE,
		Raise:     "raising",
		SendClear: true,
		Clear:     "clearing",
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestAboveRaiseLowerThanClear(t *testing.T) {
	err := Validate(&Opts[int]{
		TriggerOn: ABOVE,
		Raise:     8,
		SendClear: true,
		Clear:     10,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestAboveOrEqualRaiseEqualToClear(t *testing.T) {
	err := Validate(&Opts[int]{
		TriggerOn: ABOVE_OR_EQUAL,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestAboveOrEqualRaiseClear(t *testing.T) {
	trigger := triggerImpl[int]{
		triggerOn: ABOVE_OR_EQUAL,
		raise:     11,
		sendClear: true,
		clear:     10,
	}

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
	err := Validate(&Opts[int]{
		TriggerOn: BELOW,
		Raise:     10,
		SendClear: true,
		Clear:     9,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestBelowRaiseClear(t *testing.T) {
	trigger := triggerImpl[int]{
		triggerOn: BELOW,
		raise:     10,
		sendClear: true,
		clear:     10,
	}

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
	err := Validate(&Opts[int]{
		TriggerOn: BELOW_OR_EQUAL,
		Raise:     10,
		SendClear: true,
		Clear:     10,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestBelowOrEqualRaiseClear(t *testing.T) {
	trigger := triggerImpl[int]{
		triggerOn: BELOW_OR_EQUAL,
		raise:     100,
		sendClear: true,
		clear:     1000,
	}

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
	trigger := triggerImpl[int]{
		triggerOn: EQUAL,
		raise:     100,
		sendClear: true,
		clear:     1000,
	}

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

func TestFrom(t *testing.T) {
	trig := From[int]("ABOVE;12")
	if trig.Update(12) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trig.Update(13) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trig.Update(0) != NOTHING {
		t.Fatal("should have been NOTHING")
	}

	trig2 := From[uint]("ABOVE_OR_EQUAL;12,2")
	if trig2.Update(12) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trig2.Update(3) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trig2.Update(2) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
}

func TestNamedFrom(t *testing.T) {
	name, trig := NamedFrom[int]("ABOVE;12;mike")
	if name != "mike" {
		t.Fatal("trigger name should have been mike")
	}
	if trig.Update(12) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trig.Update(13) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trig.Update(0) != NOTHING {
		t.Fatal("should have been NOTHING")
	}

	name2, trig2 := NamedFrom[uint]("ABOVE_OR_EQUAL;12,2;molly")
	if name2 != "molly" {
		t.Fatal("trigger name should have been molly")
	}
	if trig2.Update(12) != RAISE {
		t.Fatal("should have been RAISE")
	}
	if trig2.Update(3) != NOTHING {
		t.Fatal("should have been NOTHING")
	}
	if trig2.Update(2) != CLEAR {
		t.Fatal("should have been CLEAR")
	}
}
