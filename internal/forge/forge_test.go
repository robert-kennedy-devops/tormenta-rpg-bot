package forge

import "testing"

func TestAttemptSuccess(t *testing.T) {
	out, err := Attempt(0, 0.0, 1.0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Status != OutcomeSuccess || out.NewLevel != 1 || out.Broken {
		t.Fatalf("unexpected outcome: %#v", out)
	}
}

func TestAttemptFailSafeUntilPlus4(t *testing.T) {
	out, err := Attempt(3, 0.99, 0.0) // target +4
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Status != OutcomeFailSafe || out.Broken || out.NewLevel != 3 {
		t.Fatalf("unexpected outcome: %#v", out)
	}
}

func TestAttemptCanBreakFromPlus5(t *testing.T) {
	out, err := Attempt(4, 0.99, 0.0) // target +5, fail + break roll
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Status != OutcomeBroken || !out.Broken {
		t.Fatalf("expected broken outcome, got %#v", out)
	}
}
