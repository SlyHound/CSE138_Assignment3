package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"src/utility"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	port = ":8085"
)

func setupRouter(kvStore map[string]string) *gin.Engine {
	router := gin.Default()
	socket_addr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")
	gin.SetMode(gin.ReleaseMode)
	// keep global variable of our SOCKET ADDRESS
	gin.DefaultWriter = ioutil.Discard
	// main functionality from assignment 2, basically need to modify the PUTS and DELETES to echo to other
	utility.PutRequest(router, kvStore, socket_addr, view)
	utility.GetRequest(router, kvStore, socket_addr, view)
	utility.DeleteRequest(router, kvStore, socket_addr, view)
	return router
}

func main() {

	var kvStore = make(map[string]string) // key-value store for PUT, GET, & DELETE requests (exported variable)

	router := setupRouter(kvStore)
	err := router.Run(port)
	if err != nil {
		fmt.Println("There was an issue attempting to start the server", err, "was returned.")
	}
}
