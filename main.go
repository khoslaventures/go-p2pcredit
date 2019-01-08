package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/oleiade/lane"
	"github.com/urfave/cli"
)

// bufSize is the size of the buffers for receiving and sending messages
const bufSize = 4096
const defaultPort = 12345
const candidate = "akash"
const trustlineLimit = 100

func startClient(host *Host) {
	// Send loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		in, _ := reader.ReadString('\n')
		in = strings.TrimSuffix(in, "\n")
		s := strings.Split(in, " ")
		switch s[0] {
		case "pay":
			// example: pay Bob 10
			if len(s) == 3 {
				peerID := s[1]
				if peer, exists := host.peerIDtoPeer[peerID]; exists {
					amt, err := strconv.ParseUint(s[2], 10, 64)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if amt > trustlineLimit {
						fmt.Printf("Err: Payment of %d exceeds trustline limit of %d\n", amt, trustlineLimit)
						continue
					}
					if !peer.pending {
						bal := peer.trustline.PeerBalance
						newBal := bal + int(amt)
						if newBal > trustlineLimit {
							// Send the amount that will increase the balance to 100
							payment := trustlineLimit - bal
							msgPay := Message{host.Name, peerID, "Pay", uint64(payment)}
							fmt.Println("Payment to trustlineLimit queued.")
							host.outbound <- &msgPay

							// Settle on the blockchain
							msgSettle := Message{host.Name, peerID, "Settle", trustlineLimit}
							fmt.Println("Settlement queued.")
							host.outbound <- &msgSettle

							// Send the remaining amount
							remainder := newBal - trustlineLimit
							amt = uint64(remainder)
						}
						msg := Message{host.Name, peerID, "Pay", amt}
						fmt.Println("Payment queued.")
						host.outbound <- &msg
					} else {
						fmt.Printf("Err: Connection with %s is waiting to be accepted.\n", peerID)
					}
				} else {
					fmt.Printf("Err: Connection with %s does not exists.\n", peerID)
				}
			}
		case "settle":
			// example: settle Bob 20
			// only the person with debt can settle. (i.e. negative balance)
			if len(s) == 3 {
				peerID := s[1]
				if peer, exists := host.peerIDtoPeer[peerID]; exists {
					amt, err := strconv.ParseUint(s[2], 10, 64)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if peer.trustline.HostBalance >= 0 {
						fmt.Printf("Err: Nothing to settle as HostBalance is %d\n", peer.trustline.HostBalance)
					} else {
						if !peer.pending {
							msg := Message{host.Name, peerID, "Settle", amt}
							fmt.Println("Settlement queued.")
							host.outbound <- &msg
						} else {
							fmt.Printf("Err: Connection with %s is waiting to be accepted.\n", peerID)
						}
					}
				} else {
					fmt.Printf("Err: Connection with %s does not exists.\n", peerID)
				}
			}
		case "propose":
			// same as open_trustline
			// example: propose Bob
			// look up PeerID, obtain connection details
			if len(s) == 2 {
				peerID := s[1]
				if _, exists := host.peerIDtoPeer[peerID]; !exists {
					ud := getUsers()
					for id, info := range ud {
						if id == peerID {
							host.createConnection(peerID, &info.PeerInfo)
							msg := Message{host.Name, peerID, "Propose", 0}
							fmt.Println("Propose queued.")
							host.outbound <- &msg
						}
					}
				} else {
					fmt.Printf("Err: Connection with %s already exists.\n", peerID)
				}
			}
		case "balance":
			displayTrustlineBalances(host)
		case "users":
			// print users on the FakeChain
			ud := getUsers()
			printPeerDetails(ud)
		case "delete":
			// delete all users on the FakeChain
			deleteUsers()
		case "y":
			if host.urgentcmd.Head() != nil {
				p := host.urgentcmd.Dequeue()
				prop := p.(*Proposal)
				prop.peer.PeerID = prop.msg.HostID
				prop.peer.trustline = &Trustline{0, 0}
				prop.peer.pending = false
				host.peerIDtoPeer[prop.msg.HostID] = prop.peer
				msg := Message{host.Name, prop.msg.HostID, "ProposeAccept", 0}
				host.outbound <- &msg
			}
		case "n":
			if host.urgentcmd.Head() != nil {
				p := host.urgentcmd.Dequeue()
				prop := p.(*Proposal)
				host.peerIDtoPeer[prop.msg.HostID] = prop.peer
				msg := Message{host.Name, prop.msg.HostID, "ProposeReject", 0}
				host.outbound <- &msg
			}
		case "exit":
			fmt.Println("TODO")
			// settle all debts
			// exit with code
		}
	}
}

func startService(name string, balance uint64, port uint16, isMainNet bool) {
	fmt.Println("Starting...")
	reader := bufio.NewReader(os.Stdin)
	queue := lane.NewQueue()
	host := Host{
		Name:         name,
		Port:         port,
		peers:        make(map[*Peer]bool),
		peerIDtoPeer: make(map[string]*Peer),
		inbound:      make(chan *Message),
		outbound:     make(chan *Message),
		proposal:     make(chan *Proposal),
		register:     make(chan *Peer),
		unregister:   make(chan *Peer),
		urgentcmd:    queue,
		Balance:      balance,
		reader:       reader,
	}

	fmt.Printf("Hi %s! We'll need a password for your Fakechain account.\n", host.Name)
	host.setPassword()
	host.setIP(isMainNet)

	// Add the user to Fakechain! TODO: Ensure success, if not, panic
	addUser(host.Name, host.Balance, host.password, host.IP, host.Port)
	fmt.Printf("User %s created and registered on FakeChain!\n", host.Name)

	pstr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", pstr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Set up complete, listening on port", pstr)

	go host.stateManager()
	go host.connectionListener(ln)
	startClient(&host)
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
		cli.BoolFlag{
			Name:  "mainnet, m",
			Usage: "Launch on mainnet. Else, defaults to local network.",
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
			fmt.Println(c.Bool("mainnet"))
			startService(name, balance, uint16(port), c.Bool("mainnet"))
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
