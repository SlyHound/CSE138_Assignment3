package utility

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

//DeleteRequest Client endpoint for deletions
func DeleteRequest(r *gin.Engine, dict map[string]StoreVal, localAddr string, view []string) {

	println(view)
	r.DELETE("/key-value-store/:key", func(c *gin.Context) {
		key := c.Param("key")

		// if the key-value pair exists, then delete it //
		if _, exists := dict[key]; exists {
			c.JSON(http.StatusOK, gin.H{"doesExist": true, "message": "Deleted successfully"})
			delete(dict, key)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"doesExist": false, "error": "Key does not exist", "message": "Error in DELETE"})
		}

		//Broadcast delete to all other replicas
		for i := 0; i < len(view); i++ {
			println("Replicating message to: " + "http://" + view[i] + "/key-value-store-r/" + key)
			c.Request.URL.Host = view[i]
			c.Request.URL.Scheme = "http"
			fwdRequest, err := http.NewRequest("DELETE", "http://"+view[i]+"/key-value-store-r/"+key, nil)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}

			fwdRequest.Header = c.Request.Header

			httpForwarder := &http.Client{}
			fwdResponse, err := httpForwarder.Do(fwdRequest)

			// Shouldn't worry about Error checking? just send requests out and if things are down oh well?
			if err != nil {
				msg := "Error in " + fwdRequest.Method
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": view[i] + " is down", "message": msg})
			}
			if fwdResponse != nil {
				body, _ := ioutil.ReadAll(fwdResponse.Body)
				rawJSON := json.RawMessage(body)
				c.JSON(fwdResponse.StatusCode, rawJSON)
				defer fwdResponse.Body.Close()
			}
		}
	})

}

//ReplicateDelete endpoint to replicate delete messages
func ReplicateDelete(r *gin.Engine, dict map[string]StoreVal, local_addr string, view []string) {
	r.DELETE("/key-value-store-r/:key", func(c *gin.Context) {
		key := c.Param("key")

		// if the key-value pair exists, then delete it //
		if _, exists := dict[key]; exists {
			c.JSON(http.StatusOK, gin.H{"doesExist": true, "message": "Deleted successfully"})
			delete(dict, key)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"doesExist": false, "error": "Key does not exist", "message": "Error in DELETE"})
		}
	})
}
