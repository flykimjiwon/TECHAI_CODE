package gitinfo

import "testing"

func TestParseBranch_TrimsNewline(t *testing.T) {
	if got := ParseBranch("main\n"); got != "main" {
		t.Errorf("ParseBranch = %q, want main", got)
	}
}

func TestParseBranch_DetachedHead(t *testing.T) {
	// `git rev-parse --abbrev-ref HEAD` prints "HEAD" when detached.
	if got := ParseBranch("HEAD\n"); got != "HEAD" {
		t.Errorf("ParseBranch = %q, want HEAD", got)
	}
}

func TestParseBranch_Empty(t *testing.T) {
	if got := ParseBranch(""); got != "" {
		t.Errorf("ParseBranch = %q, want empty", got)
	}
}

func TestParsePorcelain_CleanTree(t *testing.T) {
	info := Info{Branch: "main"}
	ApplyPorcelain(&info, "")
	if info.Dirty {
		t.Error("expected clean tree, got Dirty=true")
	}
	if info.Staged != 0 || info.Unstaged != 0 || info.Untracked != 0 {
		t.Errorf("counts = (%d,%d,%d), want (0,0,0)",
			info.Staged, info.Unstaged, info.Untracked)
	}
}

func TestParsePorcelain_Untracked(t *testing.T) {
	info := Info{Branch: "main"}
	ApplyPorcelain(&info, "?? newfile.go\n?? other.txt\n")
	if !info.Dirty {
		t.Error("Dirty = false, want true")
	}
	if info.Untracked != 2 {
		t.Errorf("Untracked = %d, want 2", info.Untracked)
	}
	if info.Staged != 0 || info.Unstaged != 0 {
		t.Errorf("Staged/Unstaged should be 0, got (%d,%d)", info.Staged, info.Unstaged)
	}
}

func TestParsePorcelain_MixedStatus(t *testing.T) {
	// Porcelain format: XY <path>
	//   X = staged, Y = unstaged
	//   'M' modified, 'A' added, 'D' deleted, '?' untracked, ' ' clean
	input := "M  staged.go\n M unstaged.go\nMM both.go\n?? new.go\nA  added.go\n"
	info := Info{Branch: "main"}
	ApplyPorcelain(&info, input)
	if !info.Dirty {
		t.Error("Dirty = false, want true")
	}
	// M  staged, MM staged+unstaged, A  staged → 3 staged
	if info.Staged != 3 {
		t.Errorf("Staged = %d, want 3", info.Staged)
	}
	// " M" unstaged, "MM" unstaged → 2 unstaged
	if info.Unstaged != 2 {
		t.Errorf("Unstaged = %d, want 2", info.Unstaged)
	}
	if info.Untracked != 1 {
		t.Errorf("Untracked = %d, want 1", info.Untracked)
	}
}

func TestLabel_Clean(t *testing.T) {
	info := Info{Branch: "main"}
	if got := info.Label(); got != "main" {
		t.Errorf("Label = %q, want main", got)
	}
}

func TestLabel_Dirty(t *testing.T) {
	info := Info{Branch: "main", Dirty: true, Unstaged: 2}
	if got := info.Label(); got != "main*" {
		t.Errorf("Label = %q, want main*", got)
	}
}

func TestLabel_Empty(t *testing.T) {
	info := Info{}
	if got := info.Label(); got != "" {
		t.Errorf("Label = %q, want empty (no repo)", got)
	}
}

func TestSummary_Clean(t *testing.T) {
	info := Info{Branch: "main"}
	got := info.Summary()
	if got == "" {
		t.Error("Summary should not be empty for clean tree")
	}
}

func TestSummary_DirtyCounts(t *testing.T) {
	info := Info{
		Branch:    "feature/x",
		Dirty:     true,
		Staged:    1,
		Unstaged:  2,
		Untracked: 3,
	}
	got := info.Summary()
	// Should include branch and each nonzero count
	for _, want := range []string{"feature/x", "1", "2", "3"} {
		if !contains(got, want) {
			t.Errorf("Summary=%q missing %q", got, want)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
