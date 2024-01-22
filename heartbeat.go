package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var mu sync.Mutex

type heartbeatBody struct {
	WorkerId string `json:"worker_id"`
}

func heatbeatHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var body heartbeatBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.IndentedJSON(400, gin.H{"error": err.Error()})
		return
	}
	workerId := body.WorkerId
	_, ok := workers[workerId]
	if !ok {
		c.IndentedJSON(400, gin.H{"error": "worker not found"})
		return
	}
	setWorkerCountdown(workerId, 3)
	setWorkerStatus(workerId, "up")
	logger.Debug("Heartbeat received from worker", zap.String("workerId", workerId))
	c.Status(200)

}

func updateCountdown() {
	mu.Lock()
	defer mu.Unlock()
	for _, v := range workers {
		v.countDown--
		if v.countDown <= 0 {
			setWorkerStatus(v.Id, "down")
		}
		setWorkerCountdown(v.Id, v.countDown)
	}
}
