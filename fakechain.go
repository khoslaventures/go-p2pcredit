package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Interacts with Fakechain by making HTTP calls.
// Expects a REST server on the other side.

const destURL = "http://ec2-34-222-59-29.us-west-2.compute.amazonaws.com:5000/"
const candidateKey = "akash"

// AddUser is for adding a user to FakeChain
type AddUser struct {
	Candidate string `json:"candidate"`
	ID        string `json:"public_key"`
	Balance   uint64 `json:"amount"`
	Password  string `json:"private_key"`
	PeerInfo  struct {
		IP   string `json:"host"`
		Port uint16 `json:"port"`
	} `json:"peering_info"`
}

// PayUser is for paying someone on FakeChain
type PayUser struct {
	Candidate string `json:"candidate"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Password  string `json:"private_key"`
	Amount    uint   `json:"amount"`
}

// GetOrDeleteUsers is used to get a JSON of users or delete all users
type GetOrDeleteUsers struct {
	Candidate string `json:"candidate"`
}

func addUser(id string, balance uint64, password string, ip string, port uint16) {
	endpoint := "add_user"
	m := AddUser{Candidate: candidate, ID: id, Balance: balance, Password: password}
	m.PeerInfo.IP = ip
	m.PeerInfo.Port = port

	b, err := json.Marshal(m)
	ferror(err)

	resp, err := http.Post(destURL+endpoint, "application/json", bytes.NewBuffer(b))
	ferror(err)

	fmt.Println(resp.Body)
}

func payUser(sender string, receiver string, password string, amount uint) {
	endpoint := "pay_user"
	m := PayUser{candidate, sender, receiver, password, amount}

	b, err := json.Marshal(m)
	ferror(err)

	resp, err := http.Post(destURL+endpoint, "application/json", bytes.NewBuffer(b))
	ferror(err)

	fmt.Println(resp.Body)
}

func getUsers() {
	endpoint := "get_users"
	m := GetOrDeleteUsers{candidateKey}
	b, err := json.Marshal(m)
	ferror(err)

	resp, err := http.Post(destURL+endpoint, "application/json", bytes.NewBuffer(b))
	ferror(err)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	ferror(err)

	fmt.Printf("Results: %v\n", data)
}

func deleteUsers() {
	endpoint := "delete_all_users"
	m := GetOrDeleteUsers{candidateKey}
	b, err := json.Marshal(m)
	ferror(err)

	resp, err := http.Post(destURL+endpoint, "application/json", bytes.NewBuffer(b))
	ferror(err)

	fmt.Println(resp.Body)
}

func ferror(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
