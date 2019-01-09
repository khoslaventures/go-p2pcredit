package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/howeyc/gopass"
)

// Print balance in each trustline and total trustline balance
func displayTrustlineBalances(host *Host) {
	totalTrustlineBalance := 0
	for id, peer := range host.peerIDtoPeer {
		fmt.Printf("%s: %d\n", id, peer.trustline.HostBalance)
		totalTrustlineBalance += peer.trustline.HostBalance
	}
	fmt.Printf("Total: %d\n", totalTrustlineBalance)
}

// GetOutboundIP gets preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func ferror(err error) {
	if err != nil {
		panic(err)
	}
}

func promptPassword() []byte {
	fmt.Printf("Password: ")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		// Handle gopass.ErrInterrupted or getch() read error
		panic(err)
	}
	return pass
}

func (host *Host) setPassword() {
	pass := promptPassword()
	host.password = string(pass)
}

func (host *Host) setIP(isLocal bool) {
	if !isLocal {
		fmt.Println("Running with public IP!")
		fmt.Printf("Getting IP address from ipify...\n")
		url := "https://api.ipify.org?format=text"

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		ip, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		host.IP = string(ip)
	} else {
		fmt.Println("Running with localhost only!")
		host.IP = "127.0.0.1"
	}
}
