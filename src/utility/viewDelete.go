package utility

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type sockAddr struct {
	Address string `json:"socket-address"`
}

/* this function deletes the replica from its own
view and the replica from all other replica's views */
func RequestDelete(allSocketAddrs []string, personalSocketAddr string, indiciesToRemove map[int]string) []string {

	for index, addr := range allSocketAddrs {
		if addr == personalSocketAddr { // skip over the personal replica since we don't send to ourselves
			continue
		} else if indiciesToRemove[index] == addr { // if the element exists in the map, then delete it //

			data := strings.NewReader(`{"socket-address":"` + allSocketAddrs[index] + `"}`)
			request, err := http.NewRequest("DELETE", "http://"+allSocketAddrs[index]+"/key-value-store-view", data)

			if err != nil {
				fmt.Println("There was an error creating a DELETE request.")
				break
			}

			request.Header.Set("Content-Type", "application/json")

			httpForwarder := &http.Client{}
			response, err := httpForwarder.Do(request)

			if err != nil { // if a response doesn't come back, then that replica might be down
				fmt.Println("There was an error sending a DELETE request to " + allSocketAddrs[index])
				continue
			}
			defer response.Body.Close()
		}
	}

	var allKeys []int

	// gets all the keys first to sort before removing all the replica's that failed to get a rqst back //
	for index := range indiciesToRemove {
		allKeys = append(allKeys, index)
	}

	sort.Ints(allKeys)
	for index := range allKeys {
		allSocketAddrs = append(allSocketAddrs[:index], allSocketAddrs[index+1:]...)
	}

	fmt.Println("Check allSocketAddrs in rqstDelete:", allSocketAddrs)
	return allSocketAddrs
}

func ResponseDelete(r *gin.Engine, view []string) {
	var d sockAddr
	r.DELETE("/key-value-store-view", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			fmt.Println("There was an error attempting to read the request body.")
			return
		}

		strBody := string(body[:])
		fmt.Println("Check strBody in respDelete:", strBody)
		json.NewDecoder(strings.NewReader(strBody)).Decode(&d)
		fmt.Println("Check d.SocketAddr in respDelete:", d.Address)
		defer c.Request.Body.Close()

		presentInView := false
		oIndex := 0

		for index, viewSocketAddr := range view {
			if d.Address == viewSocketAddr {
				presentInView = true
				oIndex = index
				break
			}
		}

		// if the passed in socket address is present in the current replica's view, then delete it, else 404 error //
		if presentInView {
			view = append(view[:oIndex], view[oIndex+1:]...) // deletes the replica from the current view that received the DELETE rqst. //
			fmt.Println("Check view in respDelete:", view)
			c.JSON(http.StatusOK, gin.H{"message": "Replica deleted successfully from the view"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Socket address does not exist in the view", "message": "Error in DELETE"})
		}
	})
}
