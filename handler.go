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
	c.JSON(http.StatusOK, gin.H{"msg": response})
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

	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func startServiceHandler(c *gin.Context) {

	worker_id := c.Param("worker_id")
	var requestBody StartOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	err := startServiceContainer(workers[worker_id], requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error starting container"})
		return
	}
	response := fmt.Sprintf("Container of service %s with of worker %s started", requestBody.ContainerName, worker_id)

	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func runServiceHandler(c *gin.Context) {

	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody RunOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	err := runService(workers[worker_id], services[service], requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error running service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of worker %s is running", service, worker_id)

	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func checkpointServiceHandler(c *gin.Context) {
	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody CheckpointOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON:"})
		return
	}

	_, err := checkpointService(worker_id, services[service], requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error checkpointing service:" + err.Error()})
		return
	}
	response := fmt.Sprintf("service %s of %s is checkpointed", service, worker_id)

	c.JSON(http.StatusOK, gin.H{"msg": response})
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

	err := migrateService(src, dest, services[service], requestBody.Copt, requestBody.Ropt, requestBody.Sopt, requestBody.Stop)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error migrating service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s migrated from %s to %s", service, src, dest)

	c.JSON(http.StatusOK, gin.H{"msg": response})
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

func removeServiceHandler(c *gin.Context) {
	worker_id := c.Param("worker_id")
	service := c.Param("service")

	err := removeService(workers[worker_id], services[service])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error removing service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of %s is removed", service, worker_id)

	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func stopServiceHandler(c *gin.Context) {
	worker_id := c.Param("worker_id")
	service := c.Param("service")

	err := stopService(workers[worker_id], services[service])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error stopping service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of %s is stopped", service, worker_id)

	c.JSON(http.StatusOK, gin.H{"msg": response})
}
