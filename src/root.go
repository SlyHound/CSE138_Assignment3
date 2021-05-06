package main

import (
	"fmt"
	"os"
	"src/utility"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	port      = ":8085"
	viewCount = 2
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
	personalSocketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")

	var viewSocketAddrs [viewCount]string // there can at most be two other views
	for index, currentView := range view {
		if currentView != personalSocketAddr {
			viewSocketAddrs[index] = currentView
		}
	}

	utility.RequestGet(viewSocketAddrs)
	utility.ResponseGet(router, view)

	err := router.Run(port)

	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
