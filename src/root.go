package main

import (
	"fmt"
	"os"
	"src/utility"

	"github.com/gin-gonic/gin"
)

const (
	port = ":8085"
)

/*
replica addresses are as follows:
10.0.0.2, 10.0.0.3, 10.0.0.4
with port number 8085
endpoint is as follows: /key-value-store-view
*/

func main() {
	// var kvStore = make(map[string]string) // key value store for PUT, GET, and DELETE requests for replicas

	router := gin.Default()
	ipAddress := os.Getenv("SOCKET_ADDRESS")

	utility.RequestGet(ipAddress)

	err := router.Run(port)
	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
