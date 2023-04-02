package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hlfshell/coppermind/internal/service"
)

type HttpAPI struct {
	service *service.Service
	router  *mux.Router
	port    string
}

func NewHttpAPI(service *service.Service, port string) *HttpAPI {
	api := &HttpAPI{
		service: service,
		router:  mux.NewRouter(),
		port:    port,
	}

	api.setupRouting()

	return api
}

func (api *HttpAPI) setupRouting() {
	chatRouter := api.router.PathPrefix("/chat").Subrouter()

	chatRouter.HandleFunc("/send", api.SendMessage).Methods("POST")
}

func (api *HttpAPI) Serve() error {
	return http.ListenAndServe(api.port, api.router)
}
