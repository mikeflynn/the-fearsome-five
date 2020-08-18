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
	cid, ok := r.URL.Query()["cid"]
	if !ok || len(cid[0]) < 1 {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	msg, ok := r.URL.Query()["msg"]
	if !ok || len(msg[0]) < 1 {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	doWait := false
	waitParam, ok := r.URL.Query()["wait"]
	if ok && len(waitParam[0]) == 1 {
		doWait = true
	}

	client, err := idx.clientByUUID(cid[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := &Cmd{
		ClientUUID: client.UUID,
		Payload:    shared.NewMessage("std", msg[0], shared.EncodingText),
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
