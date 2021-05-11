package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type get struct {
	Message string
	View    []string
}

/* this function will broadcast a GET request from one replica to all other
   replica's to ensure that they are currently up. */
func RequestGet(allSocketAddrs []string, personalSocketAddr string, endpoint string) ([]string, map[int]string) {
	var v get
	noResponseIndices := make(map[int]string)
	for index, addr := range allSocketAddrs {
		if addr == personalSocketAddr || index >= len(allSocketAddrs) { // skip over the personal replica since we don't send to ourselves
			continue
		}
		// insanity is doing something over and over again and excepting a different result each time -> me right now ): //
		fmt.Println("allSocketAddrs[index], index:", allSocketAddrs[index], index)
		request, err := http.NewRequest("GET", "http://"+allSocketAddrs[index]+endpoint, nil)

		if err != nil {
			fmt.Println("There was an error creating a GET request with the following error:", err.Error())
			break
		}

		httpForwarder := &http.Client{} // alias for DefaultClient
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + allSocketAddrs[index])
			noResponseIndices[index] = allSocketAddrs[index]
			continue
		}
		// fmt.Println("Check response.Body in RequestGet:", response.Body)
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		strBody := string(body[:])
		fmt.Println("Check strBody in RequestGet:", strBody)
		json.NewDecoder(strings.NewReader(strBody)).Decode(&v)
		// fmt.Println("Check v.View, V.Message in RequestGet:", v.View, v.Message)
		fmt.Println("Checking allSocketAddrs at end of rqstGet:", allSocketAddrs)
	}
	fmt.Println("Check the v.View is about to be returned:", v.View)
	fmt.Println("Check allSocketAddrs before returning v.View:", allSocketAddrs)
	return v.View, noResponseIndices
}

func ResponseGet(r *gin.Engine, view []string) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		view = DeleteDuplicates(view)
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": view})
	})
}

// custom function designed to get all key-value pairs of the current replica to store in the new replica's store //
func KeyValueResponse(r *gin.Engine, store map[string]string) {
	r.GET("/key-value-store-values", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "All pairs retrieved successfully", "view": store})
	})
}
