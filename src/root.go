package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"src/utility"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	port = ":8085"
)

type status struct {
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

//hmmmm why don't we just keep a FIFO queue on the receiving client again
//goroutine to dispatch and send requests out to our waiting
func dispatch() {
	//while true
	//
	//	check for any queues with elements in them in reqDispatch
	//		if queue not empty
	//		while queue is not empty:
	//			try sending requests in the queue to said client
	//			if response is good
	//				LOCK/UNLOCK around this
	//				remove from queue
	//			elif response is bad
	//				keep in queue
	//				break
	//
	//

}

func setupRouter(kvStore map[string]utility.StoreVal) *gin.Engine {
	router := gin.Default()
	socketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")
	currentVC := []int{0, 0, 0, 0}
	var socketIdx int
	fmt.Printf("%v\n", view)
	for i := 0; i < len(view); i++ {
		println(view[i])
		if view[i] == socketAddr {
			println("VIEW[i]: " + view[i])
			println("SOCKETADDR: " + socketAddr)
			socketIdx = i
			//set VCIndex to i
			//funky stuff here, may be unneeded, don't remove for now
			if i == 0 {
				view = view[1:]
			} else {
				view = remove(view, i)
			}
		}
	}
	fmt.Printf("%v\n", view)
	gin.SetMode(gin.ReleaseMode)
	// keep global variable of our SOCKET ADDRESS
	gin.DefaultWriter = ioutil.Discard
	// main functionality from assignment 2, basically need to modify the PUTS and DELETES to echo to other
	utility.PutRequest(router, kvStore, socketIdx, view, currentVC)
	utility.GetRequest(router, kvStore, socketIdx, view)
	utility.DeleteRequest(router, kvStore, socketIdx, view, currentVC)
	utility.ReplicatePut(router, kvStore, socketIdx, view, currentVC)
	utility.ReplicateDelete(router, kvStore, socketIdx, view, currentVC)
	return router
}

func main() {

	var kvStore = make(map[string]utility.StoreVal) // key-value store for PUT, GET, & DELETE requests (exported variable)
	//pass in reqDispatch to our requests so we can update it
	var reqDispatch = make(map[string]http.Request) // map of addresses and their queued/stored requests to replicate if things go down

	//go dispatch(reqDispatch, mutex)
	router := setupRouter(kvStore)
	err := router.Run(port)
	if err != nil {
		fmt.Println("There was an issue attempting to start the server", err, "was returned.")
	}
}
