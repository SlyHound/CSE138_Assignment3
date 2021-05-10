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
func healthCheck(view []string, personalSocketAddr string, kvStore map[string]string) {

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
		response := utility.RequestGet(view, viewSocketAddrs, "/key-value-store-view")
		fmt.Println("Check response received:", response)
		inReplica := false
		newReplica := ""

		for _, recvSocketAddr := range response {
			inReplica = false
			newReplica = recvSocketAddr
			for _, viewSocketAddr := range view {
				if viewSocketAddr == recvSocketAddr {
					inReplica = true
					break
				}
			}
		}

		if !inReplica && newReplica != "" { // broadcast a PUT request with the new replica to add to all replica's views
			utility.RequestPut(view, viewSocketAddrs, newReplica)
			if len(kvStore) == 0 { // if the current key-value store is empty, then we need to retrieve k-v pairs from the other replica's
				response = utility.RequestGet(view, viewSocketAddrs, "/key-value-store-values")
				fmt.Println("Check GET response on values:", response)
			}
		}
	}
}

func variousResponses(router *gin.Engine, view []string, store map[string]string) {
	utility.ResponseGet(router, view)
	utility.ResponseDelete(router, view)
	utility.ResponsePut(router, view)
	utility.KeyValueResponse(router, store)
}

func main() {
	kvStore := make(map[string]string) // key value store for PUT, GET, and DELETE requests for replicas

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	personalSocketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")

	go healthCheck(view, personalSocketAddr, kvStore)
	variousResponses(router, view, kvStore)

	err := router.Run(port)

	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
