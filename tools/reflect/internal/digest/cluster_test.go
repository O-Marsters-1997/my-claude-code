package digest_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/digest"
)

const gateError = "PreToolUse:Bash hook error: [~/.claude/hooks/gate.sh]: query first\n"

func blocked(t *testing.T, id string) digest.Digest {
	t.Helper()
	s := &script{t: t}
	s.call("Bash", map[string]any{"command": "grep x"}, gateError, true)
	return digest.Build(s.agent(id))
}

func TestClustersGroupByHookAcrossAgents(t *testing.T) {
	ds := []digest.Digest{blocked(t, "a"), blocked(t, "b")}

	clusters := digest.Clusters(ds)

	if len(clusters) != 1 || clusters[0].Mechanism != "~/.claude/hooks/gate.sh" || len(clusters[0].Agents()) != 2 {
		t.Fatalf("Clusters() = %+v, want one gate.sh cluster over 2 agents", clusters)
	}
	verdicts := digest.Triage(ds, 8)
	if verdicts["a"] != digest.Skip || verdicts["b"] != digest.Skip {
		t.Errorf("Triage() = %v, want both skipped once clustered", verdicts)
	}
}

func TestClustersIgnoreSingleAgentMechanism(t *testing.T) {
	ds := []digest.Digest{blocked(t, "a")}

	if got := digest.Clusters(ds); len(got) != 0 {
		t.Errorf("Clusters() = %+v, want none", got)
	}
	if v := digest.Triage(ds, 8); v["a"] != digest.Review {
		t.Errorf("Triage() = %v, want a reviewed", v)
	}
}

func TestWasteFlags(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &script{t: t}
	s.call("Bash", map[string]any{"command": "ls"}, "ok", false)
	s.call("Bash", map[string]any{"command": "sleep 90"}, "ok", false)
	s.call("Bash", map[string]any{"command": "go test ./..."}, "ok", false)
	for i, secs := range []int{0, 1, 2, 92, 200, 270} {
		s.lines[i].Timestamp = at.Add(time.Duration(secs) * time.Second).Format(time.RFC3339Nano)
	}

	got := tagsAt(digest.Build(s.agent("a")))

	want := map[int][]string{3: {digest.Wait}, 5: {digest.Idle, digest.Slow}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("tags mismatch (-want +got):\n%s", diff)
	}
}
