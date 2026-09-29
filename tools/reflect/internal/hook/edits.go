package hook

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

func (h handler) editsDir() string {
	sid := filepath.Base(h.p.SessionID)
	if h.p.SessionID == "" || sid == "." || sid == string(filepath.Separator) {
		return ""
	}
	dir := filepath.Join(h.store.Dir(), "edits", sid)
	if os.MkdirAll(dir, 0o755) != nil {
		return ""
	}
	return dir
}

func (h handler) logEdit(e logstore.Event) {
	dir := h.editsDir()
	if dir == "" {
		_ = h.store.Append(e)
		return
	}
	key := hash12(e.File)
	first := filepath.Join(dir, key+".first")
	switch err := stash(dir, first, e); {
	case err == nil:
		return
	case !errors.Is(err, fs.ErrExist):
		_ = h.store.Append(e)
		return
	}
	if f, err := os.OpenFile(filepath.Join(dir, key+".flushed"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644); err == nil {
		_ = f.Close()
		h.flushFirst(first)
	}
	_ = h.store.Append(e)
}

func (h handler) flushFirst(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var e logstore.Event
	if json.Unmarshal(b, &e) == nil {
		_ = h.store.Append(e)
	}
}

func stash(dir, path string, e logstore.Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "stash-*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, werr := tmp.Write(b)
	if err := errors.Join(werr, tmp.Close()); err != nil {
		return err
	}
	return os.Link(tmp.Name(), path)
}
