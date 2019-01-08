package main

import "fmt"

func (host *Host) messageHandler() {
	for {
		select {
		case msg := <-host.incoming:
			fmt.Println(msg)
			// update local state
			//     switch msg.Type {
			//     case "Propose":
			//     case "Pay":
			//     case "Settle":
			//     }
			// case msg := <-host.outgoing:
			//     // update local state
			//     switch msg.Type {
			//     case "Propose":
			//     case "Pay":
			//     case "Settle":
			//     }
			// case cmd := <-host.cmd:
			//     switch cmd.Type {
			//     case "trustlines":
			//     case "users":
			//     case "delete":
			//     case "exit":
			//     case "settle":
			//     }
		}

	}
}
