package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	keyLimit int = 50 // maximum number of characters allowed for a key
)

type StoreVal struct {
	Value          string `json:"value"`
	CausalMetadata []int  `json:"causal-metadata"`
}

//PutRequest for client interaction
func PutRequest(r *gin.Engine, dict map[string]StoreVal, localAddr int, view []string) {
	var d StoreVal
	//receive request
	r.PUT("/key-value-store/:key", func(c *gin.Context) {
		key := c.Param("key")
		body, _ := ioutil.ReadAll(c.Request.Body)
		strBody := string(body[:])
		println("BODY: " + strBody)
		//hmmmm
		json.Unmarshal(body, &d)
		fmt.Printf("%v\n", d.CausalMetadata)
		defer c.Request.Body.Close()
		if strBody == "{}" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Value is missing", "message": "Error in PUT"})
		} else if len(key) > keyLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Key is too long", "message": "Error in PUT"})
		} else {
			// if a key-value pair already exists, then replace the old value //
			// TO-DO: implement causal consistency and compare causal-metadata here
			if _, exists := dict[key]; exists {
				//Causal CHECK @Jackie
				dict[key] = StoreVal{d.Value, d.CausalMetadata}
				c.JSON(http.StatusOK, gin.H{"message": "Updated successfully", "replaced": true, "causal-metadata": d.CausalMetadata})
			} else { // otherwise we insert a new key-value pair //
				dict[key] = StoreVal{d.Value, d.CausalMetadata}
				c.JSON(http.StatusCreated, gin.H{"message": "Added successfully", "replaced": false, "causal-metadata": d.CausalMetadata})
			}
		}
		//send replicas PUT as well
		for i := 0; i < len(view); i++ {
			//TODO
			//refactor to skip vs remove in VC
			//Causal INCREMENT @Jackie
			println("Replicating message to: " + "http://" + view[i] + "/key-value-store-r/" + key)
			c.Request.URL.Host = view[i]
			c.Request.URL.Scheme = "http"
			d.CausalMetadata[3] = localAddr //Index of sender address
			data := &StoreVal{Value: d.Value, CausalMetadata: d.CausalMetadata}
			jsonData, _ := json.Marshal(data)
			fwdRequest, err := http.NewRequest("PUT", "http://"+view[i]+"/key-value-store-r/"+key, bytes.NewBuffer(jsonData))
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}

			fwdRequest.Header = c.Request.Header

			httpForwarder := &http.Client{}
			fwdResponse, err := httpForwarder.Do(fwdRequest)
			_ = fwdResponse

			// Shouldn't worry about Error checking? just send requests out and if things are down oh well?
			//TODO
			//USE THIS CATCH TO SEE IF SERVER IS DOWN AND UPDATE IN VIEW @Alex
			// if err != nil {
			// 	msg := "Error in " + fwdRequest.Method
			// 	c.JSON(http.StatusServiceUnavailable, gin.H{"error": view[i] + " is down", "message": msg})
			// }
			// if fwdResponse != nil {
			// 	body, _ := ioutil.ReadAll(fwdResponse.Body)
			// 	rawJSON := json.RawMessage(body)
			// 	c.JSON(fwdResponse.StatusCode, rawJSON)
			// 	defer fwdResponse.Body.Close()
			// }
		}

	})
}

//ReplicatePut Endpoint for replication
func ReplicatePut(r *gin.Engine, dict map[string]StoreVal, localAddr int, view []string) {
	var d StoreVal
	r.PUT("/key-value-store-r/:key", func(c *gin.Context) {
		key := c.Param("key")
		body, _ := ioutil.ReadAll(c.Request.Body)
		strBody := string(body[:])
		fmt.Printf("STRBODY: %s\n", strBody)
		json.Unmarshal(body, &d)
		fmt.Printf("VALUE: %s\n", d.Value)
		defer c.Request.Body.Close()
		if strBody == "{}" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Value is missing", "message": "Error in PUT"})
		} else if len(key) > keyLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Key is too long", "message": "Error in PUT"})
		} else {
			// if a key-value pair already exists, then replace the old value //
			// TO-DO: implement causal consistency and compare causal-metadata here
			if _, exists := dict[key]; exists {
				dict[key] = StoreVal{d.Value, d.CausalMetadata}
				c.JSON(http.StatusOK, gin.H{"message": "Updated successfully", "replaced": true, "causal-metadata": d.CausalMetadata})
			} else { // otherwise we insert a new key-value pair //
				dict[key] = StoreVal{d.Value, d.CausalMetadata}
				c.JSON(http.StatusCreated, gin.H{"message": "Added successfully", "replaced": false, "causal-metadata": d.CausalMetadata})
			}
		}
	})
}