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

/* this function deletes the replica from its own
view and the replica from all other replica's views */
func RequestDelete(viewSocketAddrs []string, indexToRemove int) {
	for index := range viewSocketAddrs {

		data := strings.NewReader(`{"socket-address":` + viewSocketAddrs[indexToRemove] + `}`)
		request, err := http.NewRequest("DELETE", "http://"+viewSocketAddrs[index]+"/key-value-store-view", data)

		if err != nil {
			fmt.Println("There was an error creating a DELETE request.")
			break
		}

		request.Header.Set("Content-Type", "application/json")

		httpForwarder := &http.Client{}
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a DELETE request to " + viewSocketAddrs[index])
			defer response.Body.Close()
			continue
		}
		defer response.Body.Close()
	}
}

func ResponseDelete(r *gin.Engine, view []string) {
	var d Dict
	r.DELETE("/key-value-store-view", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			fmt.Println("There was an error attempting to read the request body.")
			c.JSON(http.StatusInternalServerError, gin.H{})
		}

		strBody := string(body[:])
		json.NewDecoder(strings.NewReader(strBody)).Decode(&d)
		defer c.Request.Body.Close()

		presentInView := false

		for _, viewSocketAddr := range view {
			if d.Value == viewSocketAddr {
				presentInView = true
				break
			}
		}

		// if the passed in socket address is present in the current replica's view, then delete it, else 404 error //
		if presentInView {
			c.JSON(http.StatusOK, gin.H{"message": "Replica deleted successfully from the view"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Socket address does not exist in the view", "message": "Error in DELETE"})
		}
	})
}
