package utility

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	keyLimit int = 50 // maximum number of characters allowed for a key
)

type Dict struct {
	Key, Value string
}

func PutRequest(r *gin.Engine, dict map[string]string, local_addr string, view []string) {
	var d Dict
	println(view)
	r.PUT("/key-value-store/:key", func(c *gin.Context) {
		key := c.Param("key")
		body, _ := ioutil.ReadAll(c.Request.Body)
		strBody := string(body[:])
		json.NewDecoder(strings.NewReader(strBody)).Decode(&d)
		defer c.Request.Body.Close()
		if strBody == "{}" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Value is missing", "message": "Error in PUT"})
		} else if len(key) > keyLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Key is too long", "message": "Error in PUT"})
		} else {
			// if a key-value pair already exists, then replace the old value //
			if _, exists := dict[key]; exists {
				dict[key] = d.Value
				c.JSON(http.StatusOK, gin.H{"message": "Updated successfully", "replaced": true})
			} else { // otherwise we insert a new key-value pair //
				dict[key] = d.Value
				c.JSON(http.StatusCreated, gin.H{"message": "Added successfully", "replaced": false})
			}
		}
	})

	for i := 0; i < len(view); i++ {
		if view[i] == local_addr {
			continue
		} else {
			handleRequests(view[i])
		}
	}
}
