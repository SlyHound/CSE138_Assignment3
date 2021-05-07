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
		/* If a request returns with a view having # of replicas > current view
		   then broadcast a PUT request (this means a replica has been added to the system) */
		response := utility.RequestGet(view, viewSocketAddrs)
		response = strings.Trim(response, "[]")
		receivedView := strings.Split(response, ",")
		inReplica := false
		newReplica := ""

		for _, recvSocketAddr := range receivedView {
			inReplica = false
			newReplica = recvSocketAddr
			for _, viewSocketAddr := range view {
				if viewSocketAddr == recvSocketAddr {
					inReplica = true
					break
				}
			}
		}

		if !inReplica { // broadcast a PUT request with the new replica to add to all replica's views
			utility.RequestPut(view, viewSocketAddrs, newReplica)
		}
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
