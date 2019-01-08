package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-querystring/query"
)

// Interacts with Fakechain by making HTTP calls.
// Expects a REST server on the other side.
// May want to do something with a session that's kept alive?

const destURL = "http://ec2-34-222-59-29.us-west-2.compute.amazonaws.com:5000/"
const candidateKey = "akash"

// AddUser is for adding a user to FakeChain
type AddUser struct {
	Candidate string `url:"candidate"`
	ID        string `url:"public_key"`
	Balance   uint64 `url:"amount"`
	Password  string `url:"private_key"`
	PeerInfo  string `url:"peering_info"`
}

// PeerInfo is for connection information on peers
type PeerInfo struct {
	IP   string `json:"host"`
	Port uint16 `json:"port"`
}

// PeerDetails is used to serialize the data from getUsers
type PeerDetails struct {
	Balance  uint64   `json:"amount"`
	PeerInfo PeerInfo `json:"peering_info"`
}

// PayUser is for paying someone on FakeChain
type PayUser struct {
	Candidate string `url:"candidate"`
	Sender    string `url:"sender"`
	Receiver  string `url:"receiver"`
	Password  string `url:"private_key"`
	Amount    uint64 `url:"amount"`
}

// GetOrDeleteUsers is used to get a JSON of users or delete all users
type GetOrDeleteUsers struct {
	Candidate string `url:"candidate"`
}

// TODO: Handle incorrect password in the CLI, not here
func addUser(id string, balance uint64, password string, ip string, port uint16) string {
	endpoint := "add_user?"
	p := PeerInfo{IP: ip, Port: port}
	pb, err := json.Marshal(p)
	ferror(err)
	m := AddUser{Candidate: candidate, ID: id, Balance: balance, Password: password, PeerInfo: string(pb)}

	v, err := query.Values(m)
	ferror(err)

	url := destURL + endpoint + v.Encode()
	resp, err := http.Get(url)
	ferror(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	ferror(err)
	bodyString := string(bodyBytes)
	// Should do proper error check, but leave it
	return bodyString
}

func payUser(sender string, receiver string, password string, amount uint64) string {
	endpoint := "pay_user?"
	m := PayUser{candidate, sender, receiver, password, amount}

	v, err := query.Values(m)
	ferror(err)

	url := destURL + endpoint + v.Encode()
	resp, err := http.Get(url)
	ferror(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	ferror(err)
	bodyString := string(bodyBytes)
	// Should do proper error check, but leave it
	return bodyString
}

func getUsers() map[string]PeerDetails {
	endpoint := "get_users?"
	m := GetOrDeleteUsers{candidateKey}

	v, err := query.Values(m)
	ferror(err)

	url := destURL + endpoint + v.Encode()
	resp, err := http.Get(url)
	ferror(err)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var data map[string]PeerDetails
	err = json.Unmarshal(body, &data)
	ferror(err)

	return data
}

func printPeerDetails(data map[string]PeerDetails) {
	for id, info := range data {
		fmt.Printf("%s (%s, %d): %d\n", id, info.PeerInfo.IP, info.PeerInfo.Port, info.Balance)
	}
}

func deleteUsers() string {
	endpoint := "delete_all_users?"
	m := GetOrDeleteUsers{candidateKey}
	v, err := query.Values(m)
	ferror(err)

	url := destURL + endpoint + v.Encode()
	resp, err := http.Get(url)
	ferror(err)

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		ferror(err)
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		return bodyString
	}
	// TODO: Should error properly
	return resp.Status
}
