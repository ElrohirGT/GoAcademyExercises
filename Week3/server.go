package main

import (
	"net/http"

	"goacademy.com/week3/api/licitations"
)

type Server struct {
	Router *http.ServeMux
}

func CreateNewServer() *Server {
	return &Server{
		Router: http.NewServeMux(),
	}
}

func (self *Server) MountHandlers() {
	self.Router.HandleFunc("/{countryCode}", licitations.GETLicitationsWithCountryName)
}
