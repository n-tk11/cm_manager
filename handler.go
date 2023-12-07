package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func addWorkerHandler(c *gin.Context) {

	// Extract path variable
	addr := c.Query("addr")

	addWorker(addr)

	response := fmt.Sprintf("worker_id %s with address %s added", strconv.Itoa(worker_count), addr)

	worker_count += 1
	// Respond with the result
	c.String(http.StatusOK, response)
}

func addServiceHandler(c *gin.Context) {

	name := c.Param("name")
	image := c.Query("image")

	addService(name, image)

	response := fmt.Sprintf("service %s with image %s added", name, image)

	c.String(http.StatusOK, response)
}

func startServiceHandler(c *gin.Context) {

	worker_id := c.Param("worker_id")
	var requestBody StartOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}
	id, err := strconv.Atoi(worker_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return
	}
	startServiceContainer(workers[id], requestBody)

	response := fmt.Sprintf("Container of service %s with of worker %d started", requestBody.ContainerName, id)

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
	id, err := strconv.Atoi(worker_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return
	}
	runService(workers[id], services[service], requestBody)

	response := fmt.Sprintf("service %s with of worker %d is running", service, id)

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
	src_id, err := strconv.Atoi(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return
	}
	dest_id, err := strconv.Atoi(dest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return
	}
	migrateService(src_id, dest_id, services[service], requestBody.Copt, requestBody.Ropt, requestBody.Sopt, requestBody.Stop)

	response := fmt.Sprintf("service %s migrated from worker %s to worker %s", service, src, dest)

	c.String(http.StatusOK, response)
}
