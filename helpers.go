package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/howeyc/gopass"
)

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

func (host *Host) setIP(isMainNet bool) {
	if isMainNet {
		fmt.Println("Running on mainnet!")
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
		fmt.Printf("Your IP is: %s\n", ip)
	} else {
		fmt.Println("Running on local network!")
		host.IP = "localhost"
	}
}
