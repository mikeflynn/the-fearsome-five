package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

func adminGetList(idx *Index, w http.ResponseWriter, r *http.Request) {
	resp := idx.list()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func adminGetStatus(idx *Index, w http.ResponseWriter, r *http.Request) {
	whoami := "unknown"
	tokenInterface := r.Context().Value("user")
	if tokenInterface != nil {
		claims := tokenInterface.(*jwt.Token).Claims.(jwt.MapClaims)
		whoami = claims["username"].(string)
	}

	recentUsers := map[string]string{}
	for u, t := range idx.admins {
		recentUsers[u] = time.Now().Sub(t).String()
	}

	resp := map[string]interface{}{
		"server_version": Version,
		"client_count":   len(idx.clients),
		"uptime":         time.Now().Sub(StartTime).String(),
		"whoami":         whoami,
		"recent_users":   recentUsers,
		"settings":       Flags.params,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func adminFileReq(idx *Index, w http.ResponseWriter, r *http.Request) {
	filepath := r.FormValue("filepath")
	if filepath == "" {
		http.Error(w, "Missing filepath.", http.StatusBadRequest)
		return
	}

	// ID the target
	cid := r.FormValue("cid")
	if cid == "" {
		http.Error(w, "Missing client UUID.", http.StatusBadRequest)
		return
	}

	client, err := idx.clientByUUID(cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := &Cmd{
		ClientUUID: client.UUID,
		Payload:    shared.NewMessage("fileRequest", []byte(filepath), shared.EncodingText),
	}

	client.waitingOnResp = true
	idx.broadcast <- cmd

	resp := <-client.respChan
	client.waitingOnResp = false

	w.Header().Set("Content-Type", http.DetectContentType(resp.Payload.Body))
	if _, err := io.Copy(w, bytes.NewReader(resp.Payload.Body)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func adminFileSend(idx *Index, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // Max filesize of 32MB

	// Grab the file
	file, _, err := r.FormFile("thefile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	// Save file to buffer for transfer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileData := []byte(buf.Bytes())

	// ID the target
	cid := r.FormValue("cid")
	if cid == "" {
		http.Error(w, "Missing client UUID.", http.StatusBadRequest)
		return
	}

	client, err := idx.clientByUUID(cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := &Cmd{
		ClientUUID: client.UUID,
		Payload:    shared.NewMessage("fileTransfer", fileData, ""),
	}

	// Did a path get set?
	filepath := r.FormValue("filepath")
	if filepath != "" {
		cmd.Payload.SetMeta("filepath", filepath)
	}

	client.waitingOnResp = true
	idx.broadcast <- cmd

	resp := <-client.respChan
	client.waitingOnResp = false

	Logger("API", string(resp.Payload.Body))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func adminRunCommand(idx *Index, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
	}

	cid := r.FormValue("cid")
	if cid == "" {
		http.Error(w, "Internal error", http.StatusBadRequest)
		return
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
		Payload:    shared.NewMessage("runCommand", []byte(msg), shared.EncodingText),
	}

	if doWait {
		client.waitingOnResp = true
	}

	idx.broadcast <- cmd

	w.Header().Set("Content-Type", "application/json")

	if doWait {
		resp := <-client.respChan
		client.waitingOnResp = false

		Logger("API", string(resp.Payload.Body))

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	} else {
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	}
}
