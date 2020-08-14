package main

/*
A namespace for cross-platform utility functions,
like getting a username or file path.
*/

import (
	//"os"
	"net"
	"os/user"
	"strings"
)

func GetThisUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return u.Username, nil
}

func GetThisUserGroups() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	gids, err := u.GroupIds()
	if err != nil {
		return "", err
	}

	groups := []string{}
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)

		if err == nil {
			groups = append(groups, g.Name)
		}
	}

	return strings.Join(groups, ","), nil
}

func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
