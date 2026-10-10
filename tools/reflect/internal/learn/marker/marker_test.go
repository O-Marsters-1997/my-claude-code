package marker_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/marker"
)

func TestParseCommentStyles(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"slash", "// LEARN(go-idiomatic): wrap errors"},
		{"hash", "# LEARN(go-idiomatic): wrap errors"},
		{"block", "/* LEARN(go-idiomatic): wrap errors */"},
		{"jsx", "{/* LEARN(go-idiomatic): wrap errors */}"},
		{"html", "<!-- LEARN(go-idiomatic): wrap errors -->"},
		{"trailing", "x := f() // LEARN(go-idiomatic): wrap errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marker.Parse([]string{tt.line, "y := 1"})
			if len(got) != 1 || got[0].Text != "wrap errors" || got[0].Skill != "go-idiomatic" || got[0].Later {
				t.Fatalf("Parse(%q) = %+v, want one fix marker with skill and text", tt.line, got)
			}
		})
	}
}

func TestParseTargets(t *testing.T) {
	lines := []string{
		"package x",
		"// LEARN: first line",
		"// second line",
		"",
		"// plain note",
		"return nil",
		"call() // LEARN later: trailing",
	}
	want := []marker.Marker{
		{Text: "first line second line", Line: 2, End: 3, Target: 6, TargetText: "return nil"},
		{Text: "trailing", Later: true, Line: 7, End: 7, Target: 7, TargetText: "call() // LEARN later: trailing"},
	}
	if diff := cmp.Diff(want, marker.Parse(lines)); diff != "" {
		t.Errorf("Parse mismatch (-want +got):\n%s", diff)
	}
}

func TestParseIgnoresNonComments(t *testing.T) {
	lines := []string{`s := "LEARN: not a comment"`, "// LEARNED: no", "// nothing"}
	if got := marker.Parse(lines); len(got) != 0 {
		t.Errorf("Parse = %+v, want none", got)
	}
}

func TestParseBlockComment(t *testing.T) {
	lines := []string{"/*", " * LEARN: block text", " * more text", " */", "code()"}
	got := marker.Parse(lines)
	if len(got) != 1 || got[0].Text != "block text more text" || got[0].Target != 5 {
		t.Errorf("Parse = %+v, want continued text targeting line 5", got)
	}
}

func TestParseCommentNoMarker(t *testing.T) {
	if _, ok := marker.ParseComment("just a note"); ok {
		t.Error("ParseComment(note) ok = true, want false")
	}
}
