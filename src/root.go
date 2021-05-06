package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"src/utility"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	port        = ":8085"
	numReplicas = 2
)

/*
replica addresses are as follows:
10.0.0.2, 10.0.0.3, 10.0.0.4
with port number 8085
endpoint is as follows: /key-value-store-view
*/

// checks to ensure that replica's are up by broadcasting GET requests //
func healthCheck(view []string, personalSocketAddr string) {
	viewSocketAddrs := make([]string, numReplicas) // there can at most be two other views
	index := 0
	for _, currentView := range view {
		if currentView != personalSocketAddr {
			viewSocketAddrs[index] = currentView
			index += 1
		}
	}

	// runs infinitely on a 1 second clock interval //
	interval := time.Tick(time.Second * 1)
	for range interval {
		utility.RequestGet(view, viewSocketAddrs)
	}
}

func variousResponses(router *gin.Engine, view []string) {
	utility.ResponseGet(router, view)
	utility.ResponseDelete(router, view)
	utility.ResponsePut(router, view)
}

func main() {
	// var kvStore = make(map[string]string) // key value store for PUT, GET, and DELETE requests for replicas

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	personalSocketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")

	go healthCheck(view, personalSocketAddr)
	variousResponses(router, view)

	err := router.Run(port)

	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
