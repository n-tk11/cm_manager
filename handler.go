package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	var requestBody workerReq
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	if _, ok := workers[requestBody.Worker_id]; ok {
		logger.Error("Worker already exists", zap.String("worker_id", requestBody.Worker_id))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Worker already exists"})
		return
	}
	addWorker(requestBody.Worker_id, requestBody.Addr)

	response := fmt.Sprintf("worker_id %s with address %s added", requestBody.Worker_id, requestBody.Addr)

	// Respond with the result
	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func addServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	var requestBody serviceReq
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}
	if _, ok := services[requestBody.Name]; ok {
		logger.Error("Service already exists", zap.String("serviceName", requestBody.Name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service already exists"})
		return
	}
	addService(requestBody.Name, requestBody.Image)

	response := fmt.Sprintf("service %s with image %s added", requestBody.Name, requestBody.Image)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func startServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody StartOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}
	if _, ok := services[service]; !ok {
		logger.Error("Service not found", zap.String("serviceName", service))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service not found"})
		return
	}
	if requestBody.Image == "" {
		requestBody.Image = services[requestBody.ContainerName].Image
	}
	err := startServiceContainer(workers[worker_id], requestBody)
	if err != nil {
		logger.Error("Error starting container", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error starting container"})
		return
	}
	response := fmt.Sprintf("Container of service %s with of worker %s started", requestBody.ContainerName, worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func runServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody RunOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	err := runService(workers[worker_id], services[service], requestBody)
	if err != nil {
		logger.Error("Error running service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error running service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of worker %s is running", service, worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func checkpointServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")
	service := c.Param("service")
	var requestBody CheckpointOptions
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}

	_, err := checkpointService(worker_id, services[service], requestBody)
	if err != nil {
		logger.Error("Error checkpointing service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error checkpointing service:" + err.Error()})
		return
	}
	response := fmt.Sprintf("service %s of %s is checkpointed", service, worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func migrateServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	service := c.Param("service")
	src := c.Query("src")
	dest := c.Query("dest")

	var requestBody MigrateBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Error decoding JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding JSON"})
		return
	}
	if requestBody.Sopt.Image == "" {
		requestBody.Sopt.Image = services[requestBody.Sopt.ContainerName].Image
	}
	err := migrateService(src, dest, services[service], requestBody.Copt, requestBody.Ropt, requestBody.Sopt, requestBody.Stop)
	if err != nil {
		logger.Error("Error migrating service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error migrating service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s migrated from %s to %s", service, src, dest)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func getAllWorkersHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	var workerArr []Worker
	for _, v := range workers {
		updateWorkerServices(v.Id)
		workerArr = append(workerArr, v)
	}
	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, workerArr)
}

func getAllServicesHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	var serviceArr []Service
	for _, v := range services {
		serviceArr = append(serviceArr, v)
	}
	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	c.JSON(http.StatusOK, serviceArr)
}

func getWorkerHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")

	if _, ok := workers[worker_id]; !ok {
		logger.Error("Worker not found", zap.String("workerID", worker_id))
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}
	updateWorkerServices(worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, workers[worker_id])
}

func getServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	service := c.Param("name")
	if _, ok := services[service]; !ok {
		logger.Error("Service not found", zap.String("serviceName", service))
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, services[service])
}

func removeServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")
	service := c.Param("service")

	err := removeService(workers[worker_id], services[service])
	if err != nil {
		logger.Error("Error removing service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error removing service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of %s is removed", service, worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func stopServiceHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	worker_id := c.Param("worker_id")
	service := c.Param("service")

	err := stopService(workers[worker_id], services[service])
	if err != nil {
		logger.Error("Error stopping service", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error stopping service:" + err.Error()})
		return
	}

	response := fmt.Sprintf("service %s of %s is stopped", service, worker_id)

	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.String("response", response), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"msg": response})
}

func getServiceConfigHandler(c *gin.Context) {
	logger.Info("request", zap.String("method", "get"), zap.String("path", c.Request.URL.Path))
	service := c.Param("name")
	if _, ok := serviceConfigs[service]; !ok {
		logger.Error("Service not found", zap.String("serviceName", service))
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}
	logger.Info("response", zap.String("method", "get"), zap.String("path", c.Request.URL.Path), zap.Int("status", http.StatusOK))
	c.JSON(http.StatusOK, serviceConfigs[service])
}
