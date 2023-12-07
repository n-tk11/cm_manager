package main

import (
	"fmt"
	"os"

	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	socketPath := "/var/run/cm_man.sock"

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
	router.POST("/cm_manager/v1.0/worker", addWorkerHandler)
	router.GET("/cm_manager/v1.0/worker", getAllWorkersHandler)
	router.GET("/cm_manager/v1.0/worker/:worker_id", getWorkerHandler)
	router.POST("/cm_manager/v1.0/service", addServiceHandler)
	router.GET("/cm_manager/v1.0/service", getAllServicesHandler)
	router.GET("/cm_manager/v1.0/service/:name", getServiceHandler)
	router.POST("/cm_manager/v1.0/start/:worker_id", startServiceHandler)
	router.POST("/cm_manager/v1.0/run/:worker_id/:service", runServiceHandler)
	router.POST("/cm_manager/v1.0/checkpoint/:worker_id/:service", checkpointServiceHandler)
	router.POST("/cm_manager/v1.0/migrate/:service", migrateServiceHandler)

	// Use the router as the handler for the server
	http.Serve(listener, router)
}
