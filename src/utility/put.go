package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

const (
	keyLimit int = 50 // maximum number of characters allowed for a key
)

type StoreVal struct {
	Value          string `json:"value"`
	CausalMetadata []int  `json:"causal-metadata"`
}

func canDeliver(senderVC []int, replicaVC []int) bool {
	// conditions for delivery:
	//      senderVC[senderslot] = replicaVC[senderslot] + 1
	//      senderVC[notsender] <= replicaVC[not sender]
	senderID := senderVC[3] // sender position in VC

	for i := 0; i < 3; i++ {
		if i == senderID && senderVC[i] != replicaVC[i]+1 {
			return false
		} else if i != senderID && senderVC[i] > replicaVC[i] {
			fmt.Println("canDeliver: WE CAN'T DELIVER!!")
			return false
		}
	}

	return true
}

func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

// calculate new VC: max(senderVC, replicaVC)
func updateVC(senderVC []int, replicaVC []int) []int {
	var newVC []int
	for i := 0; i < 3; i++ {
		fmt.Printf("UPDATING SENDERVC: %v\n", senderVC)
		fmt.Printf("UPDATING REPLICAVC: %v\n", replicaVC)
		newVC[i] = max(senderVC[i], replicaVC[i])
	}
	return newVC
}

//PutRequest for client interaction
func PutRequest(r *gin.Engine, dict map[string]StoreVal, localAddr int, view []string, currVC []int) {
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
				// TODO: find this replicas VC
				color.Cyan("VECTOR CLOCK VALUE AT INDEX [%d]: %v\n", localAddr, currVC)
				if canDeliver(d.CausalMetadata, currVC) {
					d.CausalMetadata = updateVC(d.CausalMetadata, currVC) // calculate new VC: max(senderVC, currVC)
					d.CausalMetadata[3] = localAddr                       // set current position to this replica
					color.Cyan("VECTOR CLOCK VALUE AT INDEX [%d]: %v\n", localAddr, currVC)
					dict[key] = StoreVal{d.Value, d.CausalMetadata}
					c.JSON(http.StatusOK, gin.H{"message": "Updated successfully", "replaced": true})
				} else {
					// not ready to be delivered
					// place request in fifo buffer to serve request later
				}
			} else { // otherwise we insert a new key-value pair //
				color.Cyan("VECTOR CLOCK VALUE AT INDEX [%d]: %v\n", localAddr, currVC)
				if canDeliver(d.CausalMetadata, currVC) {
					// calculate new VC: max(senderVC, currVC)
					d.CausalMetadata = updateVC(d.CausalMetadata, currVC) // calculate new VC: max(senderVC, currVC)
					d.CausalMetadata[3] = localAddr                       // set current position to this replica
					color.Cyan("VECTOR CLOCK VALUE AT INDEX [%d]: %v\n", localAddr, currVC)
					dict[key] = StoreVal{d.Value, d.CausalMetadata}
					c.JSON(http.StatusCreated, gin.H{"message": "Added successfully", "replaced": false})
				} else {
					// not ready to be delivered
					// place request in fifo buffer to serve request later
				}
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
			d.CausalMetadata[localAddr]++   // increment sender VC for send event
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

//ReplicatePut Endpoint for replication
func ReplicatePut(r *gin.Engine, dict map[string]StoreVal, localAddr int, view []string, currVC []int) {
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
				c.JSON(http.StatusOK, gin.H{"message": "Updated successfully", "replaced": true})
			} else { // otherwise we insert a new key-value pair //
				dict[key] = StoreVal{d.Value, d.CausalMetadata}
				c.JSON(http.StatusCreated, gin.H{"message": "Added successfully", "replaced": false})
			}
		}
	})
}
