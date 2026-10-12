package web

import (
	"html/template"
	"net/http"
	"strings"
)

type Server struct {
	tmpl  *template.Template
	tasks TaskStore
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.board)
	mux.HandleFunc("GET /tasks/rows", s.rows)
	mux.HandleFunc("POST /tasks", s.createTask)
	mux.HandleFunc("POST /tasks/{id}/archive", s.archive)
}

func (s *Server) board(w http.ResponseWriter, r *http.Request) {
	tasks, _ := s.tasks.List(r.Context())
	if r.Header.Get("Hx-Request") != "" && r.URL.Query().Get("part") == "rows" {
		s.tmpl.ExecuteTemplate(w, "rows", tasks)
		return
	}
	s.tmpl.ExecuteTemplate(w, "layout", map[string]any{"Tasks": tasks, "Path": r.URL.Path})
}

func (s *Server) rows(w http.ResponseWriter, r *http.Request) {
	tasks, _ := s.tasks.List(r.Context())
	s.tmpl.ExecuteTemplate(w, "rows", tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	s.tasks.Create(r.Context(), title)
	tasks, _ := s.tasks.List(r.Context())
	s.tmpl.ExecuteTemplate(w, "layout", map[string]any{"Tasks": tasks, "Path": "/"})
}

func (s *Server) archive(w http.ResponseWriter, r *http.Request) {
	t, err := s.tasks.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if t.Status == "running" {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("cannot archive a running task"))
		return
	}
	s.tasks.Archive(r.Context(), t.ID)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
