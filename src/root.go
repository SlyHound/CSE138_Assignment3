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

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func setupRouter(kvStore map[string]utility.StoreVal) *gin.Engine {
	router := gin.Default()
	socketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")
	fmt.Printf("%v\n", view)
	for i := 0; i < len(view); i++ {
		println(view[i])
		if view[i] == socketAddr {
			println("VIEW[i]: " + view[i])
			println("SOCKETADDR: " + socketAddr)
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
	utility.PutRequest(router, kvStore, socketAddr, view)
	utility.GetRequest(router, kvStore, socketAddr, view)
	utility.DeleteRequest(router, kvStore, socketAddr, view)
	utility.ReplicatePut(router, kvStore, socketAddr, view)
	utility.ReplicateDelete(router, kvStore, socketAddr, view)
	return router
}

func main() {

	var kvStore = make(map[string]utility.StoreVal) // key-value store for PUT, GET, & DELETE requests (exported variable)

	router := setupRouter(kvStore)
	err := router.Run(port)
	if err != nil {
		fmt.Println("There was an issue attempting to start the server", err, "was returned.")
	}
}
