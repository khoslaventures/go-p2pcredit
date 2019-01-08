package main

import (
	"fmt"
	"testing"
)

func TestAPI(t *testing.T) {
	// Add two users
	res := addUser("akash", 200, "password1", "localhost", 4000)
	fmt.Println(res)
	res = addUser("bob", 100, "password2", "localhost", 4001)
	fmt.Println(res)
	getUsers()

	// Akash pays bob 50
	res = payUser("akash", "bob", "password1", 50)
	fmt.Println(res)
	getUsers()

	// Delete all users
	deleteUsers()
	getUsers()
}
