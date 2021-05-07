package utility

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type View struct {
	Value string
}

/* this function will broadcast a GET request from one replica to all other
   replica's to ensure that they are currently up. */
func RequestGet(allSocketAddrs []string, viewSocketAddrs []string) string {
	var v View
	for index := range viewSocketAddrs {
		fmt.Println("viewSocketAddrs[index]:", viewSocketAddrs[index])
		request, err := http.NewRequest("GET", "http://"+viewSocketAddrs[index]+"/key-value-store-view", nil)

		if err != nil {
			fmt.Println("There was an error creating a GET request with the following error:", err.Error())
			return ""
		}

		httpForwarder := &http.Client{} // alias for DefaultClient
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + viewSocketAddrs[index])
			/* call upon RequestDelete to delete the replica from its own view and
			   broadcast to other replica's to delete that same replica from their view */
			RequestDelete(allSocketAddrs, viewSocketAddrs, index)
			continue
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		strBody := string(body[:])
		json.NewDecoder(strings.NewReader(strBody)).Decode(&v)
	}
	return v.Value
}

func ResponseGet(r *gin.Engine, view []string) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": view})
	})
}
