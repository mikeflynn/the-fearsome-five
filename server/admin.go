package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mikeflynn/the-fearsome-five/shared"
)

func adminGetList(idx *Index, w http.ResponseWriter, r *http.Request) {
	resp := idx.list()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func adminSendFile(idx *Index, w http.ResponseWriter, r *http.Request) {
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

	// Sniff and set the transfer type
	contentType := http.DetectContentType(fileData)

	fileType := shared.EncodingText
	switch {
	case contentType == "application/json":
		fileType = shared.EncodingJSON
	case strings.HasPrefix(contentType, "application"):
		fileType = shared.EncodingFile
	case strings.HasPrefix(contentType, "text"):
		fileType = shared.EncodingText
	case strings.HasPrefix(contentType, "image"):
		fileType = shared.EncodingFile
	case strings.HasPrefix(contentType, "video"):
		fileType = shared.EncodingFile
	case strings.HasPrefix(contentType, "audio"):
		fileType = shared.EncodingFile
	default:
		fileType = shared.EncodingText
	}

	cmd := &Cmd{
		ClientUUID: client.UUID,
		Payload:    shared.NewMessage("fileTransfer", fileData, fileType),
	}

	client.waitingOnResp = true

	idx.broadcast <- cmd

	resp := <-client.respChan

	client.waitingOnResp = false

	Logger("API", string(resp.Payload.Body))

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
