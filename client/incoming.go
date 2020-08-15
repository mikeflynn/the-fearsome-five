package main

import (
	"net/url"
	"runtime"
	//"github.com/go-vgo/robotgo"

	"github.com/mikeflynn/the-fearsome-five/shared"
)

func IncomingRouter(conn *shared.Conn, message string) {
	params, err := url.ParseQuery(message)
	if err != nil {
		Debug("Unable to parse message")
		return
	}

	switch mtype := params.Get("type"); mtype {
	default:
		Debug("Unknown message type.")
		Connection.Send("PONG")
	}
}

func cmdProfile() {
	profile := url.Values{}

	profile.Set("type", "client_profile")

	if username, err := GetThisUser(); err == nil {
		profile.Set("user", username)
	}

	if groups, err := GetThisUserGroups(); err == nil {
		profile.Set("groups", groups)
	}

	if ip, err := GetInternalIP(); err == nil {
		profile.Set("internal_ip", ip)
	}

	profile.Set("os", runtime.GOOS)

	Connection.Send(profile.Encode())
}
