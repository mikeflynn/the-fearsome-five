package main

import (
	"net/url"
)

func IncomingRouter(conn *Conn, message string) {
	// Parse message
	params, err := url.ParseQuery(message)
	if err != nil {
		REPLLog("Unable to parse message", 1)
		return
	}

	// Lookup parent client
	if conn.ParentID == 0 {
		REPLLog("Parent client is not set!", 1)
		return
	}

	client, err := ClientIndex.getClientByID(conn.ParentID)
	if err != nil {
		REPLLog("Parent client is not found!", 1)
		return
	}

	// Process message
	switch mtype := params.Get("type"); mtype {
	case "client_profile":
		incomingProfile(client, params)
	default:
		REPLLog("Unknown message type.", 1)
	}
}

func incomingProfile(client *Client, params url.Values) {
	REPLLog("READING PROFILE...", 0)

	client.Groups = params.Get("groups")
	client.Username = params.Get("user")
	client.OperatingSystem = params.Get("os")
	client.InternalIP = params.Get("internal_ip")
}
