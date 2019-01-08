package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// Trustline is a balance tracker between the two parties. Starts at 0 each.
type Trustline struct {
	// You may want to use a Mutex. However, with CSP, you could do a message
	// passing scheme.
	HostBalance int
	PeerBalance int
}

// Peer will hold information about the socket connection and data to be sent
type Peer struct {
	PeerID    string
	trustline *Trustline
	socket    net.Conn
	data      chan []byte
	PeerInfo  *PeerInfo
	pending   bool
}

// Host will hold all of the available peer received data and
// potential incoming and terminating peers. `chan` is used for goroutines.
type Host struct {
	Name         string
	Port         uint16
	peers        map[*Peer]bool
	peerIDtoPeer map[string]*Peer
	inbound      chan *Message
	outbound     chan *Message
	proposal     chan *Proposal
	register     chan *Peer
	unregister   chan *Peer
	Balance      uint64
	password     string
	IP           string
	reader       *bufio.Reader
}

// A Proposal is used to read the first message from the socket connection
// and set values in the peerIDtoPeer map within the stateManager, and also set
// the PeerID for a Peer.
type Proposal struct {
	peer *Peer
	msg  *Message
}

// The stateManager manages states from inbound and outbound and sends messages
// outbound.
func (host *Host) stateManager() {
	for {
		select {
		case peer := <-host.register:
			host.peers[peer] = true
			// for the createConnection case
			if peer.PeerID != "" {
				host.peerIDtoPeer[peer.PeerID] = peer
			}
		case peer := <-host.unregister:
			if _, ok := host.peers[peer]; ok {
				close(peer.data)
				delete(host.peers, peer)
				if _, ok := host.peerIDtoPeer[peer.PeerID]; ok {
					delete(host.peerIDtoPeer, peer.PeerID)
				}
			}
			peer.socket.Close()
		case prop := <-host.proposal:
			fmt.Println("Proposal Received!")
			// TODO: Potential error: Because of our looping reader, this might not
			// read anything from stdin
			fmt.Printf("%s is trying to open a trustline. Accept? [y/n]: ", prop.msg.HostID)
			in, _ := host.reader.ReadString('\n')
			fmt.Println("READ: " + in)
			if strings.TrimSpace(strings.ToLower(in)) == "y" {
				prop.peer.PeerID = prop.msg.HostID
				prop.peer.trustline = &Trustline{0, 0}
				host.peerIDtoPeer[prop.msg.HostID] = prop.peer
				// Should respond on success, send ProposeAccept/Reject?...
				msg := Message{host.Name, prop.peer.PeerID, "ProposeAccept", 0}
				host.sendData(&msg)
			} else {
				msg := Message{host.Name, prop.msg.HostID, "ProposeReject", 0}
				host.sendData(&msg)
				host.unregister <- prop.peer
				// prop.peer.socket.Close()
			}
		case msg := <-host.inbound:
			// Update local state
			// msg.HostID here will be our PeerID.
			switch msg.Type {
			case "Pay":
				if peer, ok := host.peerIDtoPeer[msg.HostID]; ok {
					peer.trustline.HostBalance += int(msg.Amount)
					peer.trustline.PeerBalance -= int(msg.Amount)
				}
			case "Settle":
				// In the real case, we should verify, but this is not real
				if peer, ok := host.peerIDtoPeer[msg.HostID]; ok {
					peer.trustline.HostBalance -= int(msg.Amount)
					peer.trustline.PeerBalance += int(msg.Amount)
				}
			case "ProposeAccept":
				peer := host.peerIDtoPeer[msg.PeerID]
				peer.pending = false
				fmt.Printf("%s has accepted your trustline request!", msg.PeerID)
			case "ProposeReject":
				peer := host.peerIDtoPeer[msg.PeerID]
				host.unregister <- peer
			}
		case msg := <-host.outbound:
			// Update local state
			switch msg.Type {
			case "Pay":
				if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
					peer.trustline.HostBalance -= int(msg.Amount)
					peer.trustline.PeerBalance += int(msg.Amount)
					host.sendData(msg)
				}
			case "Settle":
				if host.Balance > msg.Amount {
					if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
						payUser(msg.HostID, msg.PeerID, host.password, msg.Amount)
						peer.trustline.HostBalance += int(msg.Amount)
						peer.trustline.PeerBalance -= int(msg.Amount)
						host.Balance -= msg.Amount
						host.sendData(msg)
					}

				} else {
					fmt.Printf("Err: Insufficient funds to settle with %s at amount: %d\n", msg.PeerID, msg.Amount)
				}
			case "Propose":
				host.sendData(msg)
			}
		}
	}
}

func (host *Host) sendData(msg *Message) {
	mb, err := json.Marshal(msg)
	ferror(err) // should never happen
	sock := host.peerIDtoPeer[msg.PeerID].socket
	sock.Write(mb)
}

// For server to read what comes from a socket for a given Peer. This
// is ran as a goroutine. Shutsdown if invalid peer.
func (host *Host) receive(peer *Peer) {
	for {
		b := make([]byte, bufSize)
		n, err := peer.socket.Read(b)
		if err != nil {
			host.unregister <- peer
			// peer.socket.Close()
			break
		}
		if n > 0 {
			fmt.Println("RECEIVED: " + string(b))
			var msg Message
			err = json.Unmarshal(b[:n], &msg)
			if err != nil {
				fmt.Println(err)
				host.unregister <- peer
				// peer.socket.Clos()
				break
			}
			if msg.Type == "Propose" {
				prop := Proposal{peer, &msg}
				host.proposal <- &prop
			} else {
				host.inbound <- &msg
			}
		}
	}
}

func connectionListener(ln net.Listener, host *Host) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Empty PeerID until identified
		peer := &Peer{PeerID: "", socket: conn, trustline: nil, data: make(chan []byte)}
		host.register <- peer
		go host.receive(peer)
	}
}

func (host *Host) createConnection(peerID string, pi *PeerInfo) {
	var cstr string
	if pi.IP == "localhost" {
		cstr = fmt.Sprintf(":%d", pi.Port)
	} else {
		cstr = fmt.Sprintf("%s:%d", pi.IP, pi.Port)
	}
	conn, err := net.Dial("tcp", cstr)
	if err != nil {
		fmt.Println(err)
	}
	// Create peer, place in mapping
	peer := &Peer{PeerID: peerID, socket: conn, trustline: &Trustline{0, 0}, data: make(chan []byte), pending: true}
	host.register <- peer
}
