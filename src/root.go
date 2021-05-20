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
func healthCheck(view *utility.View, personalSocketAddr string, kvStore map[string]utility.StoreVal) {

	// runs infinitely on a 1 second clock interval //
	interval := time.Tick(time.Second * 1)
	for range interval {
		/* If a request returns with a view having # of replicas > current view
		   then broadcast a PUT request (this means a replica has been added to the system) */
		returnedView, noResponseIndices := utility.RequestGet(view, personalSocketAddr)
		// fmt.Println("Check response received:", returnedView, noResponseIndices)

		/* call upon RequestDelete to delete the replica from its own view and
		   broadcast to other replica's to delete that same replica from their view */
		utility.RequestDelete(view, personalSocketAddr, noResponseIndices)

		fmt.Println("Check view & returnedView in healthCheck before for:", view, returnedView)
		inReplica := false

		utility.Mu.Mutex.Lock()
		if len(returnedView) > 0 {
			for _, viewSocketAddr := range view.PersonalView {
				inReplica = false
				for _, recvSocketAddr := range returnedView {
					if viewSocketAddr == recvSocketAddr {
						inReplica = true
						break
					}
					if !inReplica {
						view.NewReplica = viewSocketAddr
						break
					}
				}
			}
		}
		utility.Mu.Mutex.Unlock()

		if view.NewReplica != "" { // broadcast a PUT request with the new replica to add to all replica's views
			// fmt.Println("Before rqstPut call")
			utility.RequestPut(view, personalSocketAddr)
			// fmt.Println("Check view in healthCheck after PUT:", view)
			if len(kvStore) == 0 { // if the current key-value store is empty, then we need to retrieve k-v pairs from the other replica's
				utility.Mu.Mutex.Lock()
				for _, addr := range view.PersonalView {
					if addr == personalSocketAddr {
						continue
					}
					dictValues := utility.KvGet(addr)
					fmt.Println("*********DICTVALUES ***********", dictValues)
					// updates the current replica's key-value store with that of the received key-value store
					for key, storeVal := range dictValues {
						_, exists := kvStore[key]
						if !exists { // if the key doesn't exist in the store, then add it
							kvStore[fmt.Sprint(key)] = utility.StoreVal{Value: storeVal.Value, CausalMetadata: storeVal.CausalMetadata}
						}
					}
				}
				utility.Mu.Mutex.Unlock()
				// fmt.Println("Check GET response on values:", dictValues)
			}
		}
	}
}

// broadcasts GET requests to ensure that all replica's have consistent key-value stores //
// func dispatch(view *utility.View, store map[string]utility.StoreVal, personalSocketAddr string) {

// 	interval := time.Tick(time.Second * 1)
// 	for range interval {
// 		// obtains all the keyvalue pairs from all other replica's //
// 		utility.Mu.Mutex.Lock()
// 		for index, addr := range view.PersonalView {
// 			if addr == personalSocketAddr {
// 				continue
// 			} else {
// 				dictValues := utility.KvGet(addr)
// 				if dictValues == nil {
// 					fmt.Printf("Replica is down!")
// 					//replica is down for some reason
// 					//mark in view that replica is down/non-responsive
// 				} else {
// 					//ensure causal consistency
// 					if store != dictValues {
// 						//no causal consistency!
// 						// x = 4, x = 3, x didn't exist
// 						// repStatus = map[string]{bool, queue}
// 						// if dictValues == nil
// 						// R1 gets kvStore v 1.0
// 						// R2 and R3 get v 1.0
// 						// R2 goes down
// 						// R1 and R3 get v 2.0
// 						// R3 goes DOWN
// 						// R1 gets V 3.0
// 						// R2 comes up
// 						// R2 gets request that violates consistency (should become v 4.0)
// 						// R1 sees that, and sends KV store v3.0 to R2
// 						// R2 gets v3.0
// 						// R1 and R2 get V 4.0
// 						// R3 comes up
// 						// R3 gets V 4.0
// 						// replica1 has x = foo (v 3.0), replica2 comes up (v 0.0), client sends request to replica2 (r2 sets version to 1.0),
// 						// with x = bar, we return 201 instead of 200 because x = foo already exists
// 					}
// 				}
// 			}

// 		}
// 		utility.Mu.Mutex.Unlock()
// 	}
// }

func variousResponses(router *gin.Engine, store map[string]utility.StoreVal, view *utility.View) {
	utility.ResponseGet(router, view)
	utility.ResponseDelete(router, view)
	utility.ResponsePut(router, view)
	utility.KeyValueResponse(router, store)
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func setupRouter(kvStore map[string]utility.StoreVal, socketAddr string, view []string, currVC []int) *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	// keep global variable of our SOCKET ADDRESS
	gin.DefaultWriter = ioutil.Discard
	var socketIdx int
	fmt.Printf("%v\n", view)
	for i := 0; i < len(view); i++ {
		println(view[i])
		if view[i] == socketAddr {
			// println("VIEW[i]: " + view[i])
			// println("SOCKETADDR: " + socketAddr)
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

	// main functionality from assignment 2, basically need to modify the PUTS and DELETES to echo to other
	utility.PutRequest(router, kvStore, socketIdx, view, currVC)
	utility.GetRequest(router, kvStore, socketIdx, view)
	utility.DeleteRequest(router, kvStore, socketIdx, view, currVC)
	utility.ReplicatePut(router, kvStore, socketIdx, view, currVC)
	utility.ReplicateDelete(router, kvStore, socketIdx, view, currVC)
	return router
}

func main() {
	var kvStore = make(map[string]utility.StoreVal) // key-value store for PUT, GET, & DELETE requests (exported variable)
	// var reqDispatch = make(map[string]http.Request) // map of addresses and their queued/stored requests to replicate if things go down

	socketAddr := os.Getenv("SOCKET_ADDRESS")
	view := strings.Split(os.Getenv("VIEW"), ",")

	currVC := []int{0, 0, 0, 0}

	v := &utility.View{}
	v.PersonalView = append(v.PersonalView, view...)
	v.NewReplica = ""

	go healthCheck(v, socketAddr, kvStore)
	// go dispatch(v, kvStore, socketAddr)

	router := setupRouter(kvStore, socketAddr, view, currVC)
	variousResponses(router, kvStore, v)

	err := router.Run(port)

	if err != nil {
		fmt.Println("There was an error attempting to run the router on port", port, "with the error", err)
	}
}
