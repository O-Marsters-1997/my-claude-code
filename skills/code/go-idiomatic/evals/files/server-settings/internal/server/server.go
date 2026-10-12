package server

import "net/http"

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) ListenAndServe(h http.Handler) error {
	srv := &http.Server{Addr: s.addr, Handler: h}
	return srv.ListenAndServe()
}
