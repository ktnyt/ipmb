package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ipfs/go-ipfs-api"

	"github.com/ktnyt/ipmb"
)

func main() {
	command := os.Args[1]

	i, err := ipmb.NewIpmb("/ip4/127.0.0.1/tcp/5001")

	if err != nil {
		log.Fatal(err)
	}

	switch command {
	case "post":
		if err := i.Post(strings.Join(os.Args[2:], " ")); err != nil {
			log.Fatal(err)
		}

		if err := i.Notify(); err != nil {
			log.Fatal(err)
		}
	case "watch":
		c := make(chan shell.PubSubRecord)
		q := make(chan bool)
		go i.Watch(c, q)
		rec := <-c
		fmt.Println(string(rec.Data()))
		q <- true
	default:
		fmt.Println("That command does not exist!")
	}
}
