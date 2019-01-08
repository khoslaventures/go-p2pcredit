package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/k0kubun/pp"
	"github.com/oleiade/lane"
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
	socket    *net.TCPConn
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
	urgentcmd    *lane.Queue
	Balance      uint32
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
			peer.socket.Close() // Maybe you don't want to close socket on unregister.
		case prop := <-host.proposal:
			// TODO: This could probably be done more seamlessly.
			fmt.Println("\nProposal Received!")
			fmt.Printf("\n%s is trying to open a trustline. Accept? [y/n]: ", prop.msg.HostID)
			host.urgentcmd.Enqueue(prop)
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
				// fmt.Println("Received ProposeAccept")
				if peer, ok := host.peerIDtoPeer[msg.HostID]; ok {
					peer.pending = false
					fmt.Printf("%s has accepted your trustline request!\n", msg.HostID)
				} else {
					fmt.Printf("Err: PeerID %s not found\n", msg.HostID)
				}
			case "ProposeReject":
				// URGENT: Looks like this is not unregistering.
				fmt.Println("Received ProposeReject")
				if peer, ok := host.peerIDtoPeer[msg.HostID]; ok {
					if _, ok := host.peers[peer]; ok {
						close(peer.data)
						delete(host.peers, peer)
						delete(host.peerIDtoPeer, msg.HostID)
						println("Deleted from host.peerIDtoPeer")
					}
					fmt.Printf("%s has rejected your trustline request!\n", msg.HostID)
				} else {
					fmt.Printf("Err: PeerID %s not found\n", msg.HostID)
				}
			}
		case msg := <-host.outbound:
			// Update local state
			switch msg.Type {
			case "Pay":
				if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
					peer.trustline.HostBalance -= int(msg.Amount)
					peer.trustline.PeerBalance += int(msg.Amount)
					peer.data <- serialize(msg)
				}
			case "Settle":
				if host.Balance > msg.Amount {
					if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
						payUser(msg.HostID, msg.PeerID, host.password, msg.Amount)
						peer.trustline.HostBalance += int(msg.Amount)
						peer.trustline.PeerBalance -= int(msg.Amount)
						host.Balance -= msg.Amount
						peer.data <- serialize(msg)
					}

				} else {
					fmt.Printf("Err: Insufficient funds to settle with %s at amount: %d\n", msg.PeerID, msg.Amount)
				}
			case "Propose":
				// fmt.Println("Sending Propose")
				if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
					peer.data <- serialize(msg)
				}
			case "ProposeAccept":
				// fmt.Println("Sending ProposeAccept")
				if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
					peer.data <- serialize(msg)
				}
			case "ProposeReject":
				fmt.Println("Sending ProposeReject")
				if peer, ok := host.peerIDtoPeer[msg.PeerID]; ok {
					peer.data <- serialize(msg)
					if _, ok := host.peers[peer]; ok {
						close(peer.data)
						delete(host.peers, peer)
						delete(host.peerIDtoPeer, msg.PeerID)
						println("Deleted from host.peerIDtoPeer")
					}
				}
			}
		}
	}
}

func (host *Host) send(peer *Peer) {
	defer peer.socket.Close()
	for {
		select {
		case mb, ok := <-peer.data:
			if !ok {
				host.unregister <- peer
				return
			}
			peer.socket.Write(mb)
		}
	}
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
			var buf bytes.Buffer
			buf.Write(b[:n])
			for buf.Len() > 0 {
				var msg Message
				d, err := buf.ReadBytes('}')
				fmt.Println(string(d))
				pp.Println(d)
				err = json.Unmarshal(d, &msg)
				if err != nil {
					fmt.Println(err)
					host.unregister <- peer
					break
				}
				if msg.Type == "Propose" {
					prop := Proposal{peer, &msg}
					host.proposal <- &prop
				} else {
					host.inbound <- &msg
				}
				fmt.Println(buf.Len())
			}
		}

	}
}

// connectionListener will wait for connections and create a receive and send
// goroutine for each peer.
func (host *Host) connectionListener(ln *net.TCPListener) {
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println(err)
		}
		// Empty PeerID until identified
		peer := &Peer{PeerID: "", socket: conn, trustline: nil, data: make(chan []byte), pending: true}
		host.register <- peer
		go host.receive(peer)
		go host.send(peer)
	}
}

// createConnection is for the host to create connections and creates a receive
// and send goroutine for the specified peer.
func (host *Host) createConnection(peerID string, pi *PeerInfo) {
	cstr := fmt.Sprintf("%s:%d", pi.IP, pi.Port)
	addr, err := net.ResolveTCPAddr("tcp", cstr)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println(err)
	}
	// Create peer, place in mapping
	peer := &Peer{PeerID: peerID, socket: conn, trustline: &Trustline{0, 0}, data: make(chan []byte), pending: true}
	host.register <- peer
	go host.receive(peer)
	go host.send(peer)
}
