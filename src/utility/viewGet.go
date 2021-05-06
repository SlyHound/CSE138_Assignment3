package utility

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* this function will broadcast a GET request from one replica to all other
   replica's to ensure that they are currently up. */
func RequestGet(viewSocketAddrs []string) {
	for index := range viewSocketAddrs {
		fmt.Println("viewSocketAddrs[index]:", viewSocketAddrs[index])
		request, err := http.NewRequest("GET", "http://"+viewSocketAddrs[index]+"/key-value-store-view", nil)

		if err != nil {
			fmt.Println("There was an error creating a GET request with the following error:", err.Error())
			return
		}

		httpForwarder := &http.Client{} // alias for DefaultClient
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + viewSocketAddrs[index])
			/* call upon RequestDelete to delete the replica from its own view and
			   broadcast to other replica's to delete that same replica from their view */
			// utility.RequestDelete(allSocketAddrs, index)
			continue
		}
		defer response.Body.Close()
	}
}

func ResponseGet(r *gin.Engine, view []string) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": view})
	})
}
