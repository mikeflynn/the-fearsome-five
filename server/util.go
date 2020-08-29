package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// Option Flags

type OptionFlags struct {
	jwtKey string
	params map[string]string
}

func InitFlags() *OptionFlags {
	addr := flag.String("listen", "0.0.0.0:8000", "API listen address.")
	verbose := flag.Bool("verbose", false, "Display extra logging.")
	sslCert := flag.String("ssl-cert", "", "Cert file for SSL.")
	sslKey := flag.String("ssl-key", "", "Key file for SSL.")
	passCode := flag.String("password", "", "Key file for SSL.")
	noAuth := flag.Bool("no-auth", false, "Turn off API auth.")

	flag.Parse()

	flags := &OptionFlags{
		jwtKey: RandStringBytesMaskImprSrcSB(12),
		params: map[string]string{},
	}

	flags.params = map[string]string{
		"addr":     *addr,
		"sslCert":  *sslCert,
		"sslKey":   *sslKey,
		"passCode": *passCode,
		"verbose":  "",
		"noAuth":   "",
	}

	if *verbose {
		flags.params["verbose"] = "1"
	}

	if *noAuth {
		flags.params["noAuth"] = "1"
	}

	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		key := strings.Split(parts[0], "_")
		if key[0] == "TFF" {
			if _, ok := flags.params[key[1]]; ok {
				flags.params[key[1]] = parts[1]
			}
		}
	}

	if flags.params["passCode"] == "" {
		flags.params["passCode"] = RandStringBytesMaskImprSrcSB(12)
	}

	return flags
}

func (f *OptionFlags) isVerbose() bool {
	if f.params["verbose"] != "" {
		return true
	}

	return false
}

func (f *OptionFlags) useAuth() bool {
	if f.params["noAuth"] != "" {
		return false
	}

	return true
}

func (f *OptionFlags) getFlag(name string) string {
	if v, ok := f.params[name]; ok {
		return v
	}

	return ""
}

// Random String

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// Network

func GetExternalIP() string {
	resp, err := http.Get("http://checkip.amazonaws.com/")
	if err != nil {
		return "0.0.0.0"
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "0.0.0.0"
	}

	return strings.TrimSpace(string(data))
}
