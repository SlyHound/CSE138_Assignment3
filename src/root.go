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

// checks to ensure that replica's are up by broadcasting GET requests //
func healthCheck(view *utility.View, personalSocketAddr string, kvStore map[string]string) {

	// runs infinitely on a 1 second clock interval //
	interval := time.Tick(time.Second * 1)
	for range interval {
		/* If a request returns with a view having # of replicas > current view
		   then broadcast a PUT request (this means a replica has been added to the system) */
		viewReceived, noResponseIndices := utility.RequestGet(view, personalSocketAddr, "/key-value-store-view")
		fmt.Println("Check response received:", viewReceived, noResponseIndices)

		/* call upon RequestDelete to delete the replica from its own view and
		   broadcast to other replica's to delete that same replica from their view */
		utility.RequestDelete(view, personalSocketAddr, noResponseIndices)

		fmt.Println("Check view in healthCheck before for:", view)
		inReplica := false
		newReplica := ""

		if len(viewReceived) > 0 {
			for _, viewSocketAddr := range view.PersonalView {
				inReplica = false
				newReplica = viewSocketAddr
				for _, recvSocketAddr := range viewReceived {
					if viewSocketAddr == recvSocketAddr {
						inReplica = true
						break
					}
				}
			}
		}

		// fmt.Println("Check view in healthCheck after for:", view)

		if !inReplica && newReplica != "" { // broadcast a PUT request with the new replica to add to all replica's views
			utility.RequestPut(view, personalSocketAddr, newReplica)
			if len(kvStore) == 0 { // if the current key-value store is empty, then we need to retrieve k-v pairs from the other replica's
				viewReceived, _ = utility.RequestGet(view, personalSocketAddr, "/key-value-store-values")
				fmt.Println("Check GET response on values:", viewReceived)
				// TODO: update the current replica's key-value store with that of the received view's
			}
		}
	}
}

func variousResponses(router *gin.Engine, store map[string]string, view *utility.View) {
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

	v := &utility.View{}
	v.PersonalView = append(v.PersonalView, view...)

	go healthCheck(v, personalSocketAddr, kvStore)
	variousResponses(router, kvStore, v)

	err := router.Run(port)

	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
