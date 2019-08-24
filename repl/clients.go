package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/url"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var (
	ClientIndex = &Index{}
)

func genClientID() int64 {
	now := time.Now().Unix()
	rand.Seed(now)
	return int64(math.Ceil(float64(time.Now().Unix() * rand.Int63n(999999) / 100000000)))
}

// Index

type Index struct {
	DatabaseFile  string
	ActiveClients map[string]*Client
}

func (this *Index) DB() (*sqlite3.Conn, error) {
	if this.DatabaseFile == "" {
		return &sqlite3.Conn{}, errors.New("DatabaseFile not set!")
	}

	database, err := sqlite3.Open("sqlite3", this.DatabaseFile)
	if err != nil {
		return &sqlite3.Conn{}, err
	}

	return database
}

func (this *Index) init() error {

	database, err := this.DB()
	if err != nil {
		return err
	}

	defer database.Close()

	statement, _ := database.Prepare(`
CREATE TABLE IF NOT EXISTS client (
	id INTEGER PRIMARY KEY,
	title TEXT,
	operating_system TEXT,
	username TEXT,
	internal_ip TEXT,
	external_ip TEXT,
	last_connection TEXT,
	created_on TEXT DEFAULT CURRENT_TIME);

CREATE TABLE IF NOT EXISTS croup (
	id INTEGER PRIMARY KEY
	name TEXT NOT NULL
	created_on TEXT DEFAULT CURRENT_TIME);

CREATE TABLE IF NOT EXISTS tag (
	id INTEGER PRIMARY KEY
	name TEXT NOT NULL
	created_on TEXT DEFAULT CURRENT_TIME);

CREATE TABLE IF NOT EXISTS client_group (
	client_id INTEGER
	group_id INTEGER
	PRIMARY KEY (client_id, group_id)
	FOREIGN KEY (client_id) REFERENCES client (id)
 	ON DELETE CASCADE ON UPDATE NO ACTION
	FOREIGN KEY (group_id) REFERENCES group (id)
 	ON DELETE CASCADE ON UPDATE NO ACTION);

CREATE TABLE IF NOT EXISTS client_tag (
	client_id INTEGER
	tag_id INTEGER
	PRIMARY KEY (client_id, tag_id)
	FOREIGN KEY (client_id) REFERENCES client (id)
 	ON DELETE CASCADE ON UPDATE NO ACTION
	FOREIGN KEY (tag_id) REFERENCES tag (id)
 	ON DELETE CASCADE ON UPDATE NO ACTION);

CREATE TABLE IF NOT EXISTS message (
	id INTEGER PRIMARY KEY
	client_id INTEGER
	payload BLOB);`)
	defer stmt.Close()
	if err = statement.Exec(); err != nil {
		fmt.Println(err)
	}
}

func (this *Index) addClient(ID int64, conn *Conn) (*Client, error) {
	if ID == 0 {
		ID = genClientID()
	}

	client := &Client{
		ID:         ID,
		Connection: conn,
	}

	if err := client.Save(); err != nil {
		return &Client{}, err
	}

	this.ActiveClients[ID] = client
	return client, nil
}

func (this *Index) getClientByID(id int64) (*Client, error) {
	for _, client := range this.Clients {
		if client.ID == id {
			return client, nil
		}
	}

	return &Client{}, errors.New("Client not found.")
}

func (this *Index) listActive() []*Client {
	return this.Clients
}

func (this *Index) filter(query string) ([]*Client, error) {
	params, err := url.ParseQuery(query)
	if err != nil {
		return []*Client{}, err
	}

	indexJSON, err := json.Marshal(this)
	if err != nil {
		return []*Client{}, err
	}

	res := map[int64]int64{}

	for n, v := range params {
		fvals := gjson.GetMany(string(indexJSON), fmt.Sprintf(`clients.#[%v="%v"].ID`, n[0], v[0]))
		fmt.Println(fvals)
		for _, x := range fvals {
			res[x.Int()] = x.Int()
		}
	}

	ret := []*Client{}
	for _, x := range res {
		c, err := this.getClientByID(x)
		if err == nil {
			ret = append(ret, c)
		}
	}

	return ret, nil
}

// Client

type Client struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	OperatingSystem string   `json:"operating_system"`
	Username        string   `json:"username"`
	InternalIP      string   `json:"internal_ip"`
	ExternalIP      string   `json:"external_ip"`
	Groups          []string `json:"groups"`
	Tags            []string `json:"tags"`
	LastConnection  int64    `json:"last_connection"`
	Connection      *Conn
}

func (this *Client) isActive() bool {
	return this.Connection.IsActive
}

func (this *Client) Save() error {
	database, err := ClientIndex.DB()
	if err != nil {
		return err
	}

	var stmt *sqlite3.Stmt
	var input []string

	if this.ID == 0 {
		stmt, err = database.Prepare(`
			INSERT INTO client
			(id, title, operating_system, username, internal_ip, external_ip, last_connection)
			VALUES
			(?, ?, ?, ?, ?, ?, ?)`)

		input = []string{this.ID, this.Title, this.OperatingSystem, this.Username, this.InternalIP, this.ExternalIP, this.LastConnection}

		if err != nil {
			return err
		}
	} else {
		stmt, err = database.Prepare(`
			UPDATE client SET
			title = ?,
			operating_system = ?,
			username = ?,
			internal_ip = ?,
			external_ip = ?,
			last_connection = ?)
			WHERE id = ?`)

		input = []string{this.Title, this.OperatingSystem, this.Username, this.InternalIP, this.ExternalIP, this.LastConnection, this.ID}

		if err != nil {
			return err
		}
	}

	defer stmt.Close()

	err = stmt.Exec(input...)
	if err != nil {
		return err
	}
}

func (this *Client) Send(message string) {
	if this.Connection {
		this.Connection.send <- []byte(message)
	}
}
