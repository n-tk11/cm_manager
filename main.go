package main

import (
	"fmt"
	"os"

	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	socketPath := "/var/run/myapi.sock"

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing socket file: %v\n", err)
		return
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Error creating socket: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Listening on Unix socket: %s\n", socketPath)

	router := gin.New()

	// Define a route with a path variable
	router.GET("/student/:name")

	// Use the router as the handler for the server
	http.Serve(listener, router)
}
