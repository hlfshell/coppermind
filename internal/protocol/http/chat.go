package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hlfshell/coppermind/pkg/chat"
)

func (api *HttpAPI) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message chat.Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.service.Chat.SendMessage(&message)
	if err != nil {
		fmt.Println("Bad result", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
