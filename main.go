package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	manager_init()
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

	logger.Info("Listening on Unix socket", zap.String("path", socketPath))

	router := gin.New()
	router.Use(cors.Default())


	// Define a route with a path variable
	router.POST("/cm_manager/v1.0/worker", addWorkerHandler)
	router.GET("/cm_manager/v1.0/worker", getAllWorkersHandler)
	router.GET("/cm_manager/v1.0/worker/:worker_id", getWorkerHandler)
	router.POST("/cm_manager/v1.0/service", addServiceHandler)
	router.GET("/cm_manager/v1.0/service", getAllServicesHandler)
	router.GET("/cm_manager/v1.0/service/:name", getServiceHandler)
	router.POST("/cm_manager/v1.0/start/:worker_id/:service", startServiceHandler)
	router.POST("/cm_manager/v1.0/run/:worker_id/:service", runServiceHandler)
	router.POST("/cm_manager/v1.0/checkpoint/:worker_id/:service", checkpointServiceHandler)
	router.POST("/cm_manager/v1.0/migrate/:service", migrateServiceHandler)
	router.DELETE("/cm_manager/v1.0/remove/:worker_id/:service", removeServiceHandler)
	router.POST("/cm_manager/v1.0/stop/:worker_id/:service", stopServiceHandler)

	go http.Serve(listener, router)

	// Create another server on a different port but use the same handler
	anotherListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		return
	}
	defer anotherListener.Close()

	logger.Info("Listening on TCP socket", zap.String("port", "8080"))

	// Use the same router as the handler for the second server
	http.Serve(anotherListener, router)
}
