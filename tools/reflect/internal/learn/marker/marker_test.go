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
		{"slash", "// REFLECT(go-idiomatic): wrap errors"},
		{"hash", "# REFLECT(go-idiomatic): wrap errors"},
		{"block", "/* REFLECT(go-idiomatic): wrap errors */"},
		{"jsx", "{/* REFLECT(go-idiomatic): wrap errors */}"},
		{"html", "<!-- REFLECT(go-idiomatic): wrap errors -->"},
		{"trailing", "x := f() // REFLECT(go-idiomatic): wrap errors"},
		{"sql", "-- REFLECT(go-idiomatic): wrap errors"},
		{"sql trailing", "select 1; -- REFLECT(go-idiomatic): wrap errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marker.Parse([]string{tt.line, "y := 1"})
			if len(got) != 1 || got[0].Text != "wrap errors" || got[0].Skill != "go-idiomatic" {
				t.Fatalf("Parse(%q) = %+v, want one marker with skill and text", tt.line, got)
			}
		})
	}
}

func TestParseTargets(t *testing.T) {
	lines := []string{
		"package x",
		"// REFLECT: first line",
		"// second line",
		"",
		"// plain note",
		"return nil",
		"call() // REFLECT: trailing",
		"// REFLECT later: not a marker",
	}
	want := []marker.Marker{
		{Text: "first line second line", Line: 2, Col: 3, End: 3, Target: 6, TargetText: "return nil"},
		{Text: "trailing", Line: 7, Col: 10, End: 7, Target: 7, TargetText: "call() // REFLECT: trailing"},
	}
	if diff := cmp.Diff(want, marker.Parse(lines)); diff != "" {
		t.Errorf("Parse mismatch (-want +got):\n%s", diff)
	}
}

func TestParseIgnoresNonComments(t *testing.T) {
	lines := []string{`s := "REFLECT: not a comment"`, "// REFLECTED: no", "// nothing"}
	if got := marker.Parse(lines); len(got) != 0 {
		t.Errorf("Parse = %+v, want none", got)
	}
}

func TestParseBlockComment(t *testing.T) {
	lines := []string{"/*", " * REFLECT: block text", " * more text", " */", "code()"}
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

func TestParseIgnoresMarkersInStringsAndCodeLines(t *testing.T) {
	lines := []string{`fmt.Println("// REFLECT: in a string")`, `// REFLECT: real`, "*p = 1", "#include <x.h>"}
	got := marker.Parse(lines)
	if len(got) != 1 || got[0].Text != "real" || got[0].Target != 3 {
		t.Errorf("Parse = %+v, want one marker targeting *p = 1", got)
	}
}
