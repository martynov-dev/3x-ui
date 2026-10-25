package version

import "testing"

func TestCompareForkVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"v3.10.2-fork.2", "3.10.2-fork.1", 1},
		{"v3.10.2-fork.2", "3.10.2-fork.2", 0},
		{"v3.10.2-fork.2", "3.10.2-fork.10", -1},
		{"v3.10.2-fork.1", "3.10.2", 1},
		{"v3.10.2", "3.10.2-fork.1", -1},
		{"v3.10.3", "3.10.2-fork.10", 1},
	} {
		if got, ok := Compare(tc.a, tc.b); !ok || got != tc.want {
			t.Errorf("Compare(%q, %q) = (%d, %v), want (%d, true)", tc.a, tc.b, got, ok, tc.want)
		}
	}
}

func TestCompareRejectsUnexpectedFormats(t *testing.T) {
	if _, ok := Compare("latest", "2.9.3"); ok {
		t.Fatal("expected non-semver latest tag to be rejected")
	}
	if _, ok := Compare("v2.9", "2.9.3"); ok {
		t.Fatal("expected short version to be rejected")
	}
	for _, value := range []string{"3.10.2-fork.", "3.10.2-fork.0", "3.10.2-fork.-1", "3.10.2-fork.1-beta", "3.10.2-beta.1"} {
		if _, ok := Compare(value, "3.10.2"); ok {
			t.Errorf("expected malformed or unsupported version %q to be rejected", value)
		}
	}
}
