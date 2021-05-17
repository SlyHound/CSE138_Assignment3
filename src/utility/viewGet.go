package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type get struct {
	Message string
	View    []string
}

type View struct {
	PersonalView []string
	NewReplica   string // pertains only to PUT requests
}

/* this function will broadcast a GET request from one replica to all other
   replica's to ensure that they are currently up. */
func RequestGet(v *View, personalSocketAddr string, endpoint string) ([]string, map[int]string) {
	var (
		g get
	)
	noResponseIndices := make(map[int]string)

	Mu.Mutex.Lock()
	fmt.Println("Check v.PersonalView before for in RqstGet:", v.PersonalView)
	for index, addr := range v.PersonalView {
		if addr == personalSocketAddr { // skip over the personal replica since we don't send to ourselves
			continue
		}
		fmt.Println("allSocketAddrs[index], index:", v.PersonalView[index], index)
		copiedViewElem := v.PersonalView[index]
		request, err := http.NewRequest("GET", "http://"+copiedViewElem+endpoint, nil)

		if err != nil {
			log.Fatal("There was an error creating a GET request with the following error:", err.Error())
		}

		Mu.Mutex.Unlock()
		httpForwarder := &http.Client{} // alias for DefaultClient
		response, err := httpForwarder.Do(request)
		Mu.Mutex.Lock()

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + v.PersonalView[index])
			noResponseIndices[index] = v.PersonalView[index]
			continue
		}
		// fmt.Println("Check response.Body in RequestGet:", response.Body)
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		strBody := string(body[:])
		// fmt.Println("Check strBody in RequestGet:", strBody)
		json.NewDecoder(strings.NewReader(strBody)).Decode(&g)
		// fmt.Println("Check v.View, V.Message in RequestGet:", v.View, v.Message)
		// fmt.Println("Checking allSocketAddrs at end of rqstGet:", v)
	}
	Mu.Mutex.Unlock()
	// fmt.Println("Check the v.View is about to be returned:", g.View)
	// fmt.Println("Check allSocketAddrs before returning v.View:", v)
	return g.View, noResponseIndices
}

func ResponseGet(r *gin.Engine, view *View) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		// view = DeleteDuplicates()
		copiedViewElem := view.PersonalView
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": copiedViewElem})
	})
}

// custom function designed to get all key-value pairs of the current replica to store in the new replica's store //
func KeyValueResponse(r *gin.Engine, store map[string]StoreVal) {
	r.GET("/key-value-store-values", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "All pairs retrieved successfully", "view": store})
	})
}
