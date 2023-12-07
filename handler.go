package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type workerReq struct {
	Worker_id string `json:"worker_id"`
	Addr      string `json:"addr"`
}

type serviceReq struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type MigrateBody struct {
	Copt CheckpointOptions `json:"copt"`
	Ropt RunOptions        `json:"ropt"`
	Sopt StartOptions      `json:"sopt"`
	Stop bool              `json:"stop"`
}

func addWorkerHandler(c *gin.Context) {

	var requestBody workerReq
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	if _, ok := workers[requestBody.Worker_id]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Worker already exists"})
		return
	}
	addWorker(requestBody.Worker_id, requestBody.Addr)

	response := fmt.Sprintf("worker_id %s with address %s added", requestBody.Worker_id, requestBody.Addr)

	// Respond with the result
	c.String(http.StatusOK, response)
}

func addServiceHandler(c *gin.Context) {

	var requestBody serviceReq
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}
	if _, ok := services[requestBody.Name]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service already exists"})
		return
	}
	addService(requestBody.Name, requestBody.Image)

	response := fmt.Sprintf("service %s with image %s added", requestBody.Name, requestBody.Image)

	c.String(http.StatusOK, response)
}

func startServiceHandler(c *gin.Context) {

	worker_id := c.Param("worker_id")
	var requestBody StartOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	startServiceContainer(workers[worker_id], requestBody)

	response := fmt.Sprintf("Container of service %s with of worker %s started", requestBody.ContainerName, worker_id)

	c.String(http.StatusOK, response)
}

func runServiceHandler(c *gin.Context) {

	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody RunOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	runService(workers[worker_id], services[service], requestBody)

	response := fmt.Sprintf("service %s of worker %s is running", service, worker_id)

	c.String(http.StatusOK, response)
}

func checkpointServiceHandler(c *gin.Context) {
	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody CheckpointOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	checkpointService(worker_id, services[service], requestBody)

	response := fmt.Sprintf("service %s of worker %s is checkpointed", service, worker_id)

	c.String(http.StatusOK, response)
}

func migrateServiceHandler(c *gin.Context) {

	service := c.Param("service")
	src := c.Query("src")
	dest := c.Query("dest")

	var requestBody MigrateBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	migrateService(src, dest, services[service], requestBody.Copt, requestBody.Ropt, requestBody.Sopt, requestBody.Stop)

	response := fmt.Sprintf("service %s migrated from worker %s to worker %s", service, src, dest)

	c.String(http.StatusOK, response)
}

func getAllWorkersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, workers)
}

func getAllServicesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, services)
}

func getWorkerHandler(c *gin.Context) {
	worker_id := c.Param("worker_id")

	if _, ok := workers[worker_id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}
	c.JSON(http.StatusOK, workers[worker_id])
}

func getServiceHandler(c *gin.Context) {
	service := c.Param("name")
	if _, ok := services[service]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}
	if _, ok := services[service]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}
	c.JSON(http.StatusOK, services[service])
}
