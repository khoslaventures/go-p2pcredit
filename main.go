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
		// TODO: There's probably a better way to render the "> " after a
		// println. Perhaps some kind of log channel would do.
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
					amt, err := strconv.ParseUint(s[2], 10, 32)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if amt > trustlineLimit {
						fmt.Printf("Err: Payment of %d exceeds trustline limit of %d\n", amt, trustlineLimit)
						continue
					}
					if !peer.pending {
						// TODO: Messages are bundling up, need to fix
						bal := peer.trustline.PeerBalance
						newBal := bal + int(amt)
						if newBal > trustlineLimit {
							// Send the amount that will increase the balance to 100
							payment := trustlineLimit - bal
							msgPay := Message{host.Name, peerID, "Pay", uint32(payment)}
							fmt.Printf("Payment of %d (trustlineLimit) with %s queued\n", trustlineLimit, peerID)
							host.outbound <- &msgPay

							// Settle on the blockchain
							msgSettle := Message{host.Name, peerID, "Settle", trustlineLimit}
							fmt.Printf("Settlement with %s queued\n", peerID)
							host.outbound <- &msgSettle

							// Send the remaining amount
							remainder := newBal - trustlineLimit
							amt = uint64(remainder)
						}
						msg := Message{host.Name, peerID, "Pay", uint32(amt)}
						fmt.Printf("Payment of %d with %s queued\n", amt, peerID)
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
					amt, err := strconv.ParseUint(s[2], 10, 32)
					if err != nil {
						fmt.Println(err)
						continue
					}
					if peer.trustline.HostBalance >= 0 {
						fmt.Printf("Err: Nothing to settle as HostBalance is %d - your peer must settle!\n", peer.trustline.HostBalance)
					} else {
						if !peer.pending {
							msg := Message{host.Name, peerID, "Settle", uint32(amt)}
							fmt.Printf("Settlement with %s queued\n", peerID)
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
							err := host.createConnection(peerID, &info.PeerInfo)
							if err != nil {
								fmt.Println(err)
							} else {
								msg := Message{host.Name, peerID, "Propose", 0}
								fmt.Println("Propose queued.")
								host.outbound <- &msg
							}
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
			fmt.Println("Exiting...")
			bal := int(host.Balance)
			// TODO: Could prevent the host from accumulating more debt than it
			// can handle by keeping track of trustline debt in the host, and
			// then preventing pays
			// For now, we just settle what we can if the trustline debt exceeds
			// what we have on Fakechain
			for id, peer := range host.peerIDtoPeer {
				if peer.trustline.HostBalance < 0 {
					if bal-peer.trustline.PeerBalance > 0 {
						msg := Message{host.Name, id, "Settle", uint32(peer.trustline.PeerBalance)}
						fmt.Printf("Settlement with %s queued\n", id)
						host.outbound <- &msg
						bal -= peer.trustline.PeerBalance
					} else {
						fmt.Printf("Err: Insufficient balance to settle with %s\n", id)
					}
				}
			}
			// TODO: Should notify the other party to shutdown, perhaps create
			// new message type of SettleClose?
			os.Exit(1)
		default:
			fmt.Println("Command options:")
			fmt.Println("pay <peerID> <amount> - pays peerID the amount in a trustline")
			fmt.Println("settle <peerID> <amount> - settles amount on Fakechain with peerID for trustline")
			fmt.Println("propose <peerID> - proposes a trustline to peerID")
			fmt.Println("balance - displays peerID and corresponding trustline balance")
			fmt.Println("users - query Fakechain for user information")
			fmt.Println("exit - settle as much debt as possible and exit")
			fmt.Println("delete - deletes all users")
		}
	}
}

func startService(name string, balance uint32, port uint16, isLocal bool) {
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
	host.setIP(isLocal)

	addUser(host.Name, host.Balance, host.password, host.IP, host.Port)
	fmt.Printf("User %s created and registered on FakeChain!\n", host.Name)

	var ip string
	if !isLocal {
		ip = "0.0.0.0"
	} else {
		ip = host.IP
	}

	cstr := fmt.Sprintf("%s:%d", ip, host.Port)
	addr, err := net.ResolveTCPAddr("tcp", cstr)
	if err != nil {
		panic(err)
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Set up complete, listening on " + cstr)

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
			Name:  "local, l",
			Usage: "enable localhost connections only",
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.NArg() >= 2 {
			name := c.Args().Get(0)
			balstr := c.Args().Get(1)
			balance, err := strconv.ParseUint(balstr, 10, 32)
			if err != nil {
				return err
			}
			if port > 0xFFFF {
				return fmt.Errorf("port number %d is too high, should be below 65536", port)
			}

			startService(name, uint32(balance), uint16(port), c.Bool("local"))
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
