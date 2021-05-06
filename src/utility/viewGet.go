package utility

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequestGet(viewSocketAddrs [2]string) {
	// attempting to create & send a GET request to one of the replica's //
	for index := range viewSocketAddrs {
		request, err := http.NewRequest("GET", "http://"+fmt.Sprint(viewSocketAddrs[index])+"/key-value-store-view", nil)

		if err != nil {
			fmt.Println("There was an error creating a GET request.")
		}

		httpForwarder := &http.Client{}
		response, err := httpForwarder.Do(request)

		if err != nil { // if a response doesn't come back, then that replica might be down
			fmt.Println("There was an error sending a GET request to " + fmt.Sprint(viewSocketAddrs[0]))
			defer response.Body.Close()
			continue
		}
		defer response.Body.Close()
		break
	}
}

func ResponseGet(r *gin.Engine, view []string) {
	r.GET("/key-value-store-view", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "View retrieved successfully", "view": view})
	})
}
