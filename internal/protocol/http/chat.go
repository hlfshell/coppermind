package http

import (
	"encoding/json"
	"net/http"

	"github.com/hlfshell/coppermind/internal/chat"
)

func (api *HttpAPI) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message chat.Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
