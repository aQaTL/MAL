package main

import (
	"github.com/aqatl/mal/mal"
	"flag"
)

var username = flag.String("u", "", "username")
var password = flag.String("p", "", "password")

func main() {
	flag.Parse()
	_ = mal.NewClient(*username, *password)

}
