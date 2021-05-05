package utility

import (
	"fmt"
	"net/http"
)

func RequestGet(ipAddress string) {
	// currently don't know what to do with the response message, so I just left it as _
	_, err := http.NewRequest("GET", "http://"+fmt.Sprint(ipAddress)+"/key-value-store-view", nil)
	if err != nil {
		fmt.Println("There was an error sending a GET request to read another replica's view.")
	}
}

func ResponseGet() {

}
