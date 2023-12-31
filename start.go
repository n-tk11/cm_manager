package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/docker/docker/api/types/mount"
)

func startServiceContainer(worker Worker, startBody StartOptions) error {
	logger.Debug("Starting service", zap.String("service", startBody.ContainerName))
	if _, ok := services[startBody.ContainerName]; ok {
		url := "http://" + worker.IpAddrPort + "/cm_controller/v1/start"
		reqJson := startBody
		reqJson.Mounts = append(reqJson.Mounts, mount.Mount{Source: "chkfs", Target: "/checkpointfs", Type: "volume"})
		requestBody, err := json.Marshal(reqJson)
		if err != nil {
			logger.Error("Error marshalling JSON", zap.Error(err))
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			logger.Error("Error creating request", zap.Error(err))
			return err
		}
		req.Close = true
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		logger.Debug("Sending request to controller", zap.String("url", url))
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error sending the request", zap.Error(err))
			return err
		}
		logger.Debug("Request sent to controller")
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading the responseBody", zap.Error(err))
			return err
		}

		if resp.StatusCode == 200 {
			logger.Info("Service's container started", zap.String("worker", worker.Id), zap.String("service", startBody.ContainerName))
			addRunService(worker.Id, ServiceInWorker{Name: startBody.ContainerName, Status: "running"})
			return nil
		} else {
			logger.Error("Start service's container fail at worker", zap.String("worker", worker.Id), zap.String("service", startBody.ContainerName), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
			return fmt.Errorf("start container fail at worker with response code %d", resp.StatusCode)

		}
	} else {
		logger.Error("Service not found", zap.String("service", startBody.ContainerName))
		return errors.New("Service not found")
	}
}
