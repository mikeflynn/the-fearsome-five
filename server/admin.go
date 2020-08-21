package main

import (
	"encoding/json"
	"net/http"

	"github.com/mikeflynn/the-fearsome-five/shared"
)

func adminGetList(idx *Index, w http.ResponseWriter, r *http.Request) {
	resp := idx.list()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func adminPostSend(idx *Index, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	cid := r.FormValue("cid")
	if cid == "" {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	if action == "" {
		action = "std"
	}

	msg := r.FormValue("msg")
	if msg == "" {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	doWait := false
	waitParam := r.FormValue("wait")
	if waitParam != "" {
		doWait = true
	}

	client, err := idx.clientByUUID(cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := &Cmd{
		ClientUUID: client.UUID,
		Payload:    shared.NewMessage(action, msg, shared.EncodingText),
	}

	if doWait {
		client.waitingOnResp = true
	}

	idx.broadcast <- cmd

	if doWait {
		resp := <-client.respChan
		client.waitingOnResp = false

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	} else {
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	}
}
