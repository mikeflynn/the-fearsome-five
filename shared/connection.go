package shared

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 60 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline        = []byte{'\n'}
	space          = []byte{' '}
	Logger         func(string)
	MaxMessageSize = 1024 * 10
)

// Connection

type Conn struct {
	Ws            *websocket.Conn
	SendChan      chan *Message
	ReadCallback  func(*Conn, *Message)
	CloseCallback func(*Conn)
	State         int // 0 = ready; -1 = closed
}

func InitConnection() *Conn {
	return &Conn{
		SendChan: make(chan *Message, 256),
		State:    -1,
	}
}

func (c *Conn) Write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

func (c *Conn) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		c.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			if !ok {
				c.Write(websocket.CloseMessage, []byte{})
				return
			}

			c.Ws.SetWriteDeadline(time.Now().Add(writeWait))

			msgType := websocket.TextMessage
			if message.Encoding == EncodingFile {
				msgType = websocket.BinaryMessage
			}

			w, err := c.Ws.NextWriter(msgType)
			if err != nil {
				return
			}

			w.Write(message.Serialize())
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			Logger("Sending ping...")
			if err := c.Write(websocket.PingMessage, nil); err != nil {
				Logger("PING ERROR: " + err.Error())
				return
			}
		}
	}
}

func (c *Conn) ReadPump() {
	defer func() {
		c.Close()
	}()

	c.Ws.SetReadLimit(int64(MaxMessageSize))
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				Logger(fmt.Sprintf("ERROR: %v", err))

				c.Close()
			}

			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		Logger("Incoming: " + string(message))

		if c.ReadCallback != nil {
			c.ReadCallback(c, ReadMessage(message))
		}
	}
}

func (c *Conn) Establish(host string, certPath string) bool {
	if c.State == -1 {
		Logger("Connecting to " + host + "...")

		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		if certPath != "" {
			certFile, err := ioutil.ReadFile(certPath)
			if err != nil {
				Logger(err.Error())
				return false
			}

			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM(certFile)
			tlsConfig = &tls.Config{RootCAs: caPool}
		}

		d := websocket.Dialer{TLSClientConfig: tlsConfig}
		ws, _, err := d.Dial(host, nil)
		if err == nil {
			Logger("Connection established!")
			c.SetWS(ws)

			c.Ws.SetCloseHandler(func(code int, text string) error {
				Logger("Closing connection...")
				c.State = -1
				return errors.New(text)
			})

			return true
		} else {
			Logger("Connection Error: " + err.Error())
		}
	}

	return false
}

func (c *Conn) SetWS(ws *websocket.Conn) {
	c.Ws = ws
	c.State = 0
}

func (c *Conn) Close() {
	if c.State == -1 {
		return
	}

	c.State = -1

	Logger("Closing connection...")

	defer func() {
		if r := recover(); r != nil {
			Logger(fmt.Sprintf("Caught panic: %v", r))
		}
	}()

	if c.CloseCallback != nil {
		c.CloseCallback(c)
	}

	err := c.Ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		Logger("Write Close Error: " + err.Error())
	}

	if err := c.Ws.Close(); err != nil {
		Logger("Error Closing Connection: " + err.Error())
	}
}

func (c *Conn) Send(message *Message) {
	c.SendChan <- message
}

// Message

type Message struct {
	Action   string            `json:"action"`
	Body     []byte            `json:"body"`
	Encoding string            `json:"encoding"`
	Metadata map[string]string `json:"metadata"`
}

const (
	EncodingText = "text/plain"
	EncodingJSON = "application/json"
	EncodingFile = "application/octet-stream"
)

func GetEncoding(input []byte) string {
	contentType := http.DetectContentType(input)

	fileType := EncodingText
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		fileType = EncodingJSON
	case strings.HasPrefix(contentType, "application"):
		fileType = EncodingFile
	case strings.HasPrefix(contentType, "text"):
		fileType = EncodingText
	case strings.HasPrefix(contentType, "image"):
		fileType = EncodingFile
	case strings.HasPrefix(contentType, "video"):
		fileType = EncodingFile
	case strings.HasPrefix(contentType, "audio"):
		fileType = EncodingFile
	default:
		fileType = EncodingText
	}

	return fileType
}

func ReadMessage(jsonStr []byte) *Message {
	m := &Message{}
	json.Unmarshal(jsonStr, m)

	return m
}

func NewMessage(action string, body []byte, encoding string) *Message {
	if encoding == "" {
		encoding = GetEncoding(body)
	}

	return &Message{
		Action:   action,
		Body:     body,
		Encoding: encoding,
		Metadata: map[string]string{},
	}
}

func (m *Message) Serialize() []byte {
	ret, _ := json.Marshal(m)
	return ret
}

func (m *Message) SetMeta(key string, value string) error {
	m.Metadata[key] = value
	return nil
}

func (m *Message) DelMeta(key string) error {
	if _, ok := m.Metadata[key]; ok {
		delete(m.Metadata, key)
		return nil
	}

	return errors.New("Metadata key not found.")
}

func (m *Message) GetMeta(key string) (string, error) {
	if val, ok := m.Metadata[key]; ok {
		return val, nil
	}

	return "", errors.New("Metadata key not found.")
}

func (m *Message) GetBodyJSON() (map[string]string, error) {
	body := map[string]string{}

	if m.Encoding != EncodingJSON {
		return body, errors.New("Body encoding is not JSON.")
	}

	err := json.Unmarshal(m.Body, &body)
	if err != nil {
		return body, err
	}

	return body, nil
}
