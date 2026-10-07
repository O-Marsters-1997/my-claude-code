package session_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
)

type linked struct {
	ID, Parent, Link, ToolUseID, Brief string
	Depth                              int
}

func links(s session.Session) []linked {
	var out []linked
	for _, a := range s.Agents {
		out = append(out, linked{a.ID, a.Parent, a.Link, a.ToolUseID, a.Brief, a.Depth})
	}
	return out
}

func TestLoadLinksEveryAgent(t *testing.T) {
	f := newFixture(t)
	f.write(f.mainPath(),
		f.prompt("build the thing"),
		f.use("tu_sync", "Agent", spawn("explore the code")),
		f.result("tu_sync", "found it", false, map[string]any{"agentId": "sync1", "status": "completed"}),
		f.say("using the exploration"),
		f.use("tu_async", "Agent", spawn("review in background")),
		f.result("tu_async", "launched", false, map[string]any{"isAsync": true, "status": "async_launched", "agentId": "async1"}),
		f.use("tu_skill", "Skill", map[string]any{"skill": "code-review", "args": "medium"}),
		f.result("tu_skill", "done", false, map[string]any{"status": "forked", "agentId": "fork1"}),
		f.use("tu_wt", "Agent", map[string]any{"prompt": "implement #1 in your worktree", "isolation": "worktree"}),
		f.result("tu_wt", "launched", false, map[string]any{"isAsync": true, "status": "async_launched", "agentId": "wt1"}),
		f.notify("async1", "tu_async", "completed"),
		f.notify("wt1", "tu_wt", "completed"),
		f.say("all agents finished"),
	)
	f.write(f.agentPath("sync1"), f.prompt("explore the code"), f.say("found it"))
	f.meta("sync1", map[string]any{"agentType": "Explore", "toolUseId": "tu_sync", "spawnDepth": 1, "model": "haiku"})
	f.write(f.agentPath("async1"),
		f.prompt("review in background"),
		f.use("tu_nested", "Agent", spawn("altitude angle")),
		f.result("tu_nested", "ok", false, map[string]any{"agentId": "nested1"}),
	)
	f.meta("async1", map[string]any{"agentType": "general-purpose", "toolUseId": "tu_async", "spawnDepth": 1})
	f.write(f.agentPath("nested1"), f.prompt("altitude angle"), f.say("ok"))
	f.meta("nested1", map[string]any{"agentType": "fork", "isFork": true, "toolUseId": "tu_nested", "parentAgentId": "async1", "spawnDepth": 2})
	f.write(f.agentPath("fork1"), f.prompt("review the diff"), f.say("done"))
	f.meta("fork1", map[string]any{"agentType": "general-purpose", "spawnDepth": 1})
	worktree := "/Users/me/repo/.claude/worktrees/agent-wt1"
	f.meta("wt1", map[string]any{"agentType": "general-purpose", "toolUseId": "tu_wt", "spawnDepth": 1, "worktreePath": worktree, "spawnedWithWorktree": true})
	wtTranscript := filepath.Join(f.projects, "-Users-me-repo--claude-worktrees-agent-wt1", "99999999-0000-0000-0000-000000000000.jsonl")
	f.write(wtTranscript, f.prompt("implement #1 in your worktree"), f.say("implemented"))

	s, err := session.Load(f.projects, f.sid)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := []linked{
		{"main", "", "", "", "", 0},
		{"sync1", "main", session.LinkToolUseID, "tu_sync", "explore the code", 1},
		{"async1", "main", session.LinkToolUseID, "tu_async", "review in background", 1},
		{"nested1", "async1", session.LinkToolUseID, "tu_nested", "altitude angle", 2},
		{"fork1", "main", session.LinkAgentID, "tu_skill", "code-review medium", 1},
		{"wt1", "main", session.LinkToolUseID, "tu_wt", "implement #1 in your worktree", 1},
	}
	if diff := cmp.Diff(want, links(s)); diff != "" {
		t.Errorf("Load() agents (-want +got):\n%s", diff)
	}
	if got := s.Agent("wt1").Path; got != wtTranscript {
		t.Errorf("worktree agent Path = %q, want %q", got, wtTranscript)
	}
	if got := s.Agent("sync1").Outcome; got != "using the exploration" {
		t.Errorf("sync1 Outcome = %q, want %q", got, "using the exploration")
	}
	if got := s.Agent("async1").Outcome; got != "all agents finished" {
		t.Errorf("async1 Outcome = %q, want %q", got, "all agents finished")
	}
	if got := s.Agent("sync1").Model; got != "haiku" {
		t.Errorf("sync1 Model = %q, want haiku", got)
	}
}

func TestLoadFallsBackToPathLink(t *testing.T) {
	f := newFixture(t)
	f.write(f.mainPath(), f.prompt("hi"), f.say("hello"))
	f.write(f.agentPath("orphan1"), f.prompt("lost brief"), f.say("ok"))
	f.meta("orphan1", map[string]any{"agentType": "general-purpose", "spawnDepth": 1})

	s, err := session.Load(f.projects, f.sid)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	a := s.Agent("orphan1")
	if a == nil || a.Link != session.LinkPath || a.Parent != "main" {
		t.Fatalf("orphan1 = %+v, want Link=%q Parent=main", a, session.LinkPath)
	}
	if len(s.Warnings) == 0 {
		t.Error("Load() Warnings is empty, want a warning for the path-only link")
	}
}

func TestLoadRefusesLiveAgents(t *testing.T) {
	f := newFixture(t)
	f.write(f.mainPath(),
		f.prompt("go"),
		f.use("tu_bg", "Agent", spawn("long job")),
		f.result("tu_bg", "launched", false, map[string]any{"isAsync": true, "status": "async_launched", "agentId": "bg1"}),
		f.command("reflect"),
	)
	f.write(f.agentPath("bg1"), f.prompt("long job"))
	f.meta("bg1", map[string]any{"agentType": "general-purpose", "toolUseId": "tu_bg", "spawnDepth": 1})

	_, err := session.Load(f.projects, f.sid)
	var live *session.LiveError
	if !errors.As(err, &live) {
		t.Fatalf("Load() error = %v, want *LiveError", err)
	}
	if diff := cmp.Diff([]string{"bg1"}, live.Agents); diff != "" {
		t.Errorf("LiveError.Agents (-want +got):\n%s", diff)
	}
}

func TestLoadCutsReflectTurns(t *testing.T) {
	f := newFixture(t)
	f.write(f.mainPath(),
		f.prompt("first task"),
		f.say("worked on it"),
		f.command("reflect"),
		f.use("tu_rev_old", "Agent", spawn("review agent main")),
		f.result("tu_rev_old", "clean", false, map[string]any{"agentId": "rev_old"}),
		f.say("old reflection report"),
		f.prompt("second task"),
		f.compact(),
		f.say("worked on the second task"),
		f.command("reflect"),
		f.use("tu_rev_new", "Agent", spawn("review agent main")),
		f.result("tu_rev_new", "launched", false, map[string]any{"isAsync": true, "agentId": "rev_new"}),
	)
	for _, aid := range []string{"rev_old", "rev_new"} {
		f.write(f.agentPath(aid), f.prompt("review agent main"))
	}
	f.meta("rev_old", map[string]any{"agentType": "general-purpose", "toolUseId": "tu_rev_old", "spawnDepth": 1})
	f.meta("rev_new", map[string]any{"agentType": "general-purpose", "toolUseId": "tu_rev_new", "spawnDepth": 1})

	s, err := session.Load(f.projects, f.sid)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(s.Agents) != 1 {
		t.Errorf("Load() kept %d agents, want only main", len(s.Agents))
	}
	var texts []string
	for _, l := range s.Main().Lines {
		text, _ := l.Parts()
		texts = append(texts, text)
	}
	got := strings.Join(texts, "|")
	for _, gone := range []string{"old reflection report", "/reflect"} {
		if strings.Contains(got, gone) {
			t.Errorf("main lines contain %q, want it cut: %s", gone, got)
		}
	}
	for _, kept := range []string{"first task", "second task", "worked on the second task"} {
		if !strings.Contains(got, kept) {
			t.Errorf("main lines lack %q: %s", kept, got)
		}
	}
}

func TestLoadUnknownSession(t *testing.T) {
	_, err := session.Load(t.TempDir(), "nope")
	if !errors.Is(err, session.ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}
