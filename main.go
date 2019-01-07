package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

// bufSize is the size of the buffers for receiving and sending messages
const bufSize = 4096
const defaultPort = 12345
const candidate = "akash"

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
}

// Host will hold all of the available peer received data and
// potential incoming and terminating peers. `chan` is used for goroutines.
type Host struct {
	Name         string
	peers        map[*Peer]bool
	peerIDtoPeer map[string]*Peer
	messages     chan *MessageChan
	register     chan *Peer
	unregister   chan *Peer
	Balance      uint64
}

// MessageChan is used to wrap a Message and Peer struct for a channel.
type MessageChan struct {
	peer *Peer
	msg  *Message
}

func (host *Host) startHandler() {
	for {
		select {
		case peer := <-host.register:
			host.peers[peer] = true
			host.peerIDtoPeer[peer.PeerID] = peer
			fmt.Println("Added new peer connection!")
		case peer := <-host.unregister:
			if _, ok := host.peers[peer]; ok {
				close(peer.data)         // closes a channel
				delete(host.peers, peer) // delete this connection from host.peers
				delete(host.peerIDtoPeer, peer.PeerID)
				fmt.Println("A peer connection has terminated!")
			}
		case msgChan := <-host.messages:
			m := msgChan.msg
			p := msgChan.peer
			switch msgChan.msg.Type {
			case "Propose":
				// Assume no crashes/attacks
				p.PeerID = msgChan.msg.PeerID
				p.trustline = &Trustline{HostBalance: 0, PeerBalance: 0}
			case "Pay":
				p.trustline.HostBalance += m.Amount
				p.trustline.PeerBalance -= m.Amount
				// if this trustline exceeds 100, then settle on chain. Perhaps
				// the flow should be:
				// - Host says he wants to send money
				// - Host sees that debt will exceed -100 for him and credit
				// will exceed 100 for Peer.
				// - Host will craft messages through a loop, such that
				// for each 100 reached on HostAmount, A settle is sent
			case "Settle":
				if m.Amount > 0 && m.Amount <= p.trustline.HostBalance {
					// make call to fakechain for payment, subtract balances
					p.trustline.HostBalance -= m.Amount
					p.trustline.PeerBalance += m.Amount
				}

			default:
				// close the connection
			}
		}
	}
}

// For server to read what comes from a socket for a given Peer. This
// is ran as a goroutine. Shutsdown if invalid peer.
func (host *Host) receive(peer *Peer) {
	for {
		b := make([]byte, bufSize)
		len, err := peer.socket.Read(b)
		if err != nil {
			host.unregister <- peer
			peer.socket.Close()
			break
		}
		if len > 0 {
			fmt.Println("RECEIVED: " + string(b))
			// In here, we could handle messages, as they are done so in a
			// goroutine. Better to pass it in a message box to the server
			// goroutine.
			// But, to validate messages and interact with state, we can do that
			// here
			// What do we need to act on stuff and respond to a peer?
			// Let's say server receives a pay message.
			// First, we serialize. Then, we look up and see which Peer it is.
			// We can do this by looking at which socket was pinged, then seek
			// out the PeerID. PeerID can then be verified if desired.
			//
			// Once we've got that, PeerID will correspond to some trustline, so
			// we can create a mapping.
			// host.broadcast <- message
			m := parseRawBytes(b)
			msgChan := &MessageChan{msg: &m, peer: peer}
			host.messages <- msgChan
		}
	}
}

// Write to corresponding peer's socket, for the messages inside of
// the data channel for the peer.
func (host *Host) send(peer *Peer) {
	defer peer.socket.Close() // on return, close the socket.
	for {
		select {
		case message, ok := <-peer.data:
			if !ok {
				return
			}
			peer.socket.Write(message)
		}
	}
}

func startService(name string, balance uint64, port uint) {
	fmt.Println("Starting...")
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
	}

	host := Host{
		Name:       name,
		peers:      make(map[*Peer]bool),
		messages:   make(chan *MessageChan),
		register:   make(chan *Peer),
		unregister: make(chan *Peer),
		Balance:    balance,
	}

	go host.startHandler()
	go startConnectionListener(ln, host)

	// insert client code here
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		in := scanner.Text()
		s := strings.Split(in, " ")
		switch s[0] {
		case "pay":
			// construct payment message
		case "settle":
			// construct settlement message
		case "balance":
			// print trustline and cash balances
		case "trustlines":
			// print trustline details
		case "users":
			// print users on the FakeChain
		case "delete":
			// delete all users on the FakeChain
		case "exit":
			// settle all debts
			// exit with code
		default:
			continue
		}

	}
}

func startConnectionListener(ln net.Listener, host Host) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Empty PeerID until identified
		peer := &Peer{PeerID: "", socket: conn, trustline: nil, data: make(chan []byte)}
		host.register <- peer
		go host.receive(peer)
		// go host.send(peer)
	}
}

func (peer *Peer) receive() {
	for {
		b := make([]byte, bufSize)
		length, err := peer.socket.Read(b)
		if err != nil {
			peer.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(b))
		}
	}
}

func client() {
	// the idea is to be able to make connections and communicate with the
	// server side
	connection, error := net.Dial("tcp", "localhost:78910")
	if error != nil {
		fmt.Println(error)
	}
	peer := &Peer{PeerID: "", socket: connection}
	go peer.receive()
}

func inputListener() {
	for {
		// Prompt for input, configuration details, etc.
	}
}

func main() {
	var port uint

	app := cli.NewApp()
	app.Name = "Trustline Payment System"
	app.Usage = "Let's you do payments off-chain for Fakechain, but insecurely."
	app.Description = "Example: ./messages --port PORT_NUMBER <username> <amount>\n\n   This will start a client that accepts connetions on PORT_NUMBER with the specified username\n   and amount."
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.UintFlag{
			Name:        "port, p",
			Value:       defaultPort,
			Usage:       "`PORT_NUMBER` for receiving messages",
			Destination: &port,
		},
	}

	app.Action = func(c *cli.Context) error {
		// alternative to using flags, just two args.
		if c.NArg() >= 2 {
			name := c.Args().Get(0)
			balstr := c.Args().Get(1)
			balance, err := strconv.ParseUint(balstr, 10, 64)
			if err != nil {
				return err
			}
			if port > 0xFFFF {
				return fmt.Errorf("port number %d is too high, should be below 65536", port)
			}
			fmt.Println(name)
			fmt.Println(balance)
			fmt.Println(port)
			startService(name, balance, port)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	// startService()
}

// Idea, deny all Golang websocket connections if they don't fit the protocol.
