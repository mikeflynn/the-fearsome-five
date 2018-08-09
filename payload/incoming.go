package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"runtime"
	//"github.com/go-vgo/robotgo"
)

func IncomingRouter(conn *Conn, message string) {
	params, err := url.ParseQuery(message)
	if err != nil {
		Debug("Unable to parse message", 0)
		return
	}

	switch mtype := params.Get("type"); mtype {
	case "init":
		cmdInit(params)
	default:
		Debug("Unknown message type.", 0)
		Connection.Send("PONG")
	}
}

func cmdInit(params url.Values) {
	Debug(fmt.Sprintf("Initialized with new client id: %v", params.Get("client_id")), 0)

	if params.Get("client_id") != "" {
		fileData := []byte(params.Get("client_id"))
		err := ioutil.WriteFile(GetConfigPath(), fileData, 0600)
		if err != nil {
			Debug("Failed to write config file.", 0)
		}

		cmdProfile()
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
