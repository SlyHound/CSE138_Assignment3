package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Dict struct {
	Key, Value string
}

func RequestPut(allSocketAddrs []string, personalSocketAddr string, newSocketAddr string) []string {

	// first add the new replica to the current view //
	allSocketAddrs = append(allSocketAddrs, newSocketAddr)

	// now broadcast a PUT request to all other replica's to add it to their view's //
	data := strings.NewReader(`{"socket-address":` + newSocketAddr + `}`)
	for index, addr := range allSocketAddrs {
		if addr == personalSocketAddr || index >= len(allSocketAddrs) { // skip over the personal replica since we don't send to ourselves
			continue
		}
		request, err := http.NewRequest("PUT", "http://"+allSocketAddrs[index]+"/key-value-store-view", data)

		if err != nil {
			fmt.Println("There was an error creating a PUT request.")
			break
		}

		request.Header.Set("Content-Type", "application/json")

		httpForwarder := &http.Client{}
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a PUT request to " + allSocketAddrs[index])
			/* call upon RequestDelete to delete the replica from its own view and
			   broadcast to other replica's to delete that same replica from their view */
			// allSocketAddrs = RequestDelete(allSocketAddrs, personalSocketAddr, index)
			continue
		}
		defer response.Body.Close()
	}
	allSocketAddrs = DeleteDuplicates(allSocketAddrs)
	return allSocketAddrs
}

func ResponsePut(r *gin.Engine, view *View) {

	var d Dict
	r.PUT("/key-value-store-view", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			fmt.Println("There was an error attempting to read the request body.")
			c.JSON(http.StatusInternalServerError, gin.H{})
		}

		strBody := string(body[:])
		json.NewDecoder(strings.NewReader(strBody)).Decode(&d)
		view.PersonalView = append(view.PersonalView, d.Value) // adds the new replica to the view //
		defer c.Request.Body.Close()

		presentInView := false

		for _, viewSocketAddr := range view.PersonalView {
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
