package mlxrunner

import "testing"

const gib = 1 << 30

func TestPlanWired(t *testing.T) {
	if got, want := planWired(64*gib), 64*gib*4/5; got != want {
		t.Errorf("planWired(64 GiB) = %d, want %d", got, want)
	}
}

func TestPlanCache(t *testing.T) {
	cases := []struct {
		name            string
		modelSize, free int
		want            int
	}{
		// free=16 GiB -> wired cap 12.8 GiB.
		{"model fits leaves default", 10 * gib, 16 * gib, 0},
		{"model at cap leaves default", 12 * gib, 16 * gib, 0},
		{"model overflows tightens", 14 * gib, 16 * gib, 16 * gib / 5},
		{"unknown free leaves default", 14 * gib, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := planCache(tc.modelSize, tc.free); got != tc.want {
				t.Errorf("planCache(%d, %d) = %d, want %d", tc.modelSize, tc.free, got, tc.want)
			}
		})
	}
}

func TestPlanSnapshotBudget(t *testing.T) {
	const floor int64 = 512 << 20
	cases := []struct {
		name            string
		modelSize, free int
		want            int64
	}{
		{"unknown free keeps default", 14 * gib, 0, defaultSnapshotBudget},
		{"model overflows free floors", 14 * gib, 10 * gib, floor},
		{"comfortable caps at default", 4 * gib, 64 * gib, defaultSnapshotBudget},
		// free=16 GiB -> reserve 3.2 GiB; 16-8-3.2 = 4.8 GiB.
		{"constrained scales to host", 8 * gib, 16 * gib, int64(16*gib - 8*gib - 16*gib/5)},
		// 16-12-3.2 = 0.8 GiB, still above the floor.
		{"just above floor", 12 * gib, 16 * gib, int64(16*gib - 12*gib - 16*gib/5)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := planSnapshotBudget(tc.modelSize, tc.free); got != tc.want {
				t.Errorf("planSnapshotBudget(%d, %d) = %d, want %d", tc.modelSize, tc.free, got, tc.want)
			}
		})
	}
}
