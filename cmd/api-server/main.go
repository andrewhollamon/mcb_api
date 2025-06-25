package main

import (
	"fmt"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/api"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/memorystore"
)

func main() {

	// first setup the memory store
	memorystore.Init()

	r := api.SetupRouter()
	// Listen and Server in 0.0.0.0:8080
	err := r.Run(":8080")
	if err == nil {
		fmt.Printf("Router ended with error %v", err)
		return
	}

	// test whether the queue is reachable, and stop accepting changes if not

}
