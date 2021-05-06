package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"src/utility"
	"strings"

	"github.com/gin-gonic/gin"
)

type dict struct {
	Key, Value string
}

func RequestPut(allSocketAddrs []string, viewSocketAddrs []string, newSocketAddr string) {

	// first add the new replica to the current view //
	allSocketAddrs = append(allSocketAddrs, newSocketAddr)

	// now broadcast a PUT request to all other replica's to add it to their view's //
	data := strings.NewReader(`{"socket-address":` + newSocketAddr + `}`)
	for index := range viewSocketAddrs {
		request, err := http.NewRequest("PUT", "http://"+viewSocketAddrs[index]+"/key-value-store-view", data)

		if err != nil {
			fmt.Println("There was an error creating a PUT request.")
			break
		}

		request.Header.Set("Content-Type", "application/json")

		httpForwarder := &http.Client{}
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a PUT request to " + viewSocketAddrs[index])
			defer response.Body.Close()
			/* call upon RequestDelete to delete the replica from its own view and
			   broadcast to other replica's to delete that same replica from their view */
			utility.RequestDelete(allSocketAddrs, index)
			continue
		}
		defer response.Body.Close()
	}
}

func ResponsePut(r *gin.Engine, view []string, allSocketAddrs []string) {

	var d dict
	r.PUT("/key-value-store-view", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			fmt.Println("There was an error attempting to read the request body.")
			c.JSON(http.StatusInternalServerError, gin.H{})
		}

		strBody := string(body[:])
		json.NewDecoder(strings.NewReader(strBody)).Decode(&d)
		allSocketAddrs = append(allSocketAddrs, d.Value)
		defer c.Request.Body.Close()

		presentInView := false

		for _, viewSocketAddr := range view {
			if d.Value == viewSocketAddr {
				presentInView = true
				break
			}
		}

		if presentInView {
			c.JSON(http.StatusNotFound, gin.H{"error": "Socket address already exists in the view", "message": "Error in PUT"})
		} else {
			c.JSON(http.StatusCreated, gin.H{"message": "Replica added successfully to the view"})
		}
	})
}
