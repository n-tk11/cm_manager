package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var lastChkRun = make(map[string]bool)

func checkpointService(worker_id string, service Service, option CheckpointOptions) (string, error) {
	logger.Debug("Checkpointing service", zap.String("service", service.Name))
	url := "http://" + workers[worker_id].IpAddrPort + "/cm_controller/v1/checkpoint/" + service.Name
	currentTime := time.Now().UTC()

	// Format the time in ISO 8601 format
	iso8601Format := "2006-01-02T15:04:05Z07:00"
	iso8601Time := currentTime.Format(iso8601Format)
	option.ImgUrl = "file:/checkpointfs/" + service.Name + "/" + service.Name + "_" + worker_id + "_" + iso8601Time
	requestBody, err := json.Marshal(option)
	if err != nil {
		logger.Error("Error marshalling JSON", zap.Error(err))
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error("Error creating request", zap.Error(err))
		return "", err
	}

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	logger.Debug("Sending request to controller", zap.String("url", url))
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error sending the request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading the responseBody", zap.Error(err))
		return "", err
	}
	if resp.StatusCode == 200 {
		config := serviceConfigs[service.Name]
		config.ChkOpt = option
		serviceConfigs[service.Name] = config
		updateWorkerServices(worker_id, service.Name)
		logger.Info("Checkpoint successfully the image name", zap.String("image", option.ImgUrl))
		addCheckpointFile(service.Name, option.ImgUrl)
		lastChkRun[service.Name] = option.LeaveRun
		return option.ImgUrl, nil
	} else {
		updateWorkerServices(worker_id, service.Name)
		logger.Error("Checkpoint service fail at worker", zap.String("worker", worker_id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("checkpoint service fail at worker with response code %d", resp.StatusCode)

	}
}
