package rpcsdk

import (
	"net/http"
)

type Server struct {
	header http.Header
}

func NewServer(header http.Header) *Server {
	return &Server{
		header: header,
	}
}

func (server *Server) GetAppId() string {
	return server.header.Get(SelfAppidKey)
}

func (server *Server) GetAccountId() string {
	return server.header.Get(AccountIdKey)
}

func (server *Server) GetFromAppId() string {
	return server.header.Get(FromAppidKey)
}

func (server *Server) GetTraceId() string {
	return server.header.Get(RequestTraceKey)
}

func (server *Server) GetHeader(key string) string {
	return server.header.Get(key)
}
