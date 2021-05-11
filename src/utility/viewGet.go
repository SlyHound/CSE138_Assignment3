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

type View struct {
	PersonalView []string
}

/* this function will broadcast a GET request from one replica to all other
   replica's to ensure that they are currently up. */
func RequestGet(v *View, personalSocketAddr string, endpoint string) ([]string, map[int]string) {
	var g get
	noResponseIndices := make(map[int]string)
	for index, addr := range v.PersonalView {
		if addr == personalSocketAddr || index >= len(v.PersonalView) { // skip over the personal replica since we don't send to ourselves
			continue
		}
		fmt.Println("allSocketAddrs[index], index:", v.PersonalView[index], index)
		request, err := http.NewRequest("GET", "http://"+v.PersonalView[index]+endpoint, nil)

		if err != nil {
			fmt.Println("There was an error creating a GET request with the following error:", err.Error())
			break
		}

		httpForwarder := &http.Client{} // alias for DefaultClient
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + v.PersonalView[index])
			noResponseIndices[index] = v.PersonalView[index]
			continue
		}
		// fmt.Println("Check response.Body in RequestGet:", response.Body)
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		strBody := string(body[:])
		fmt.Println("Check strBody in RequestGet:", strBody)
		json.NewDecoder(strings.NewReader(strBody)).Decode(&g)
		// fmt.Println("Check v.View, V.Message in RequestGet:", v.View, v.Message)
		fmt.Println("Checking allSocketAddrs at end of rqstGet:", v)
	}
	fmt.Println("Check the v.View is about to be returned:", g.View)
	fmt.Println("Check allSocketAddrs before returning v.View:", v)
	return g.View, noResponseIndices
}

func ResponseGet(r *gin.Engine, view *View) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		// view = DeleteDuplicates()
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": view.PersonalView})
	})
}

// custom function designed to get all key-value pairs of the current replica to store in the new replica's store //
func KeyValueResponse(r *gin.Engine, store map[string]string) {
	r.GET("/key-value-store-values", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "All pairs retrieved successfully", "view": store})
	})
}
