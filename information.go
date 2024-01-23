package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func updateWorkerServices(worker_id string) error {
	worker, ok := workers[worker_id]
	if !ok {
		return errors.New("worker not found")
	}
	for _, v := range worker.Services {
		status, err := queryServiceStatus(worker_id, v.Name)
		if err != nil {
			deleteRunService(worker_id, v.Name)
			continue
		}
		v.Status = status
		updateRunService(worker_id, v)

	}
	return nil
}

func queryServiceStatus(worker_id string, service string) (string, error) {
	_, ok := workers[worker_id]
	if !ok {
		return "", errors.New("worker not found")
	}
	url := "http://" + workers[worker_id].IpAddrPort + "/cm_controller/v1/service/" + service

	req, err := http.NewRequest("GET", url, nil)
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
	logger.Debug("Request sent to controller")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading response body", zap.Error(err))
			return "", err
		}
		bodyString := string(bodyBytes)
		logger.Debug("Response from controller", zap.String("body", bodyString))
		var response map[string]interface{}
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			logger.Error("Error unmarshalling response body", zap.Error(err))
			return "", err
		}
		if _, ok := response["status"]; ok {
			return response["status"].(string), nil
		}
		return "", errors.New("status not found")
	}
	logger.Error("Error getting status from controller", zap.Int("status", resp.StatusCode))
	return "", errors.New("Error getting status from controller")
}

func isWorkerUp(worker_id string) bool {
	url := "http://" + workers[worker_id].IpAddrPort + "/cm_controller/v1/up"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Error creating request", zap.Error(err))
		return false
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	logger.Debug("Sending request to controller", zap.String("url", url))
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error sending the request", zap.Error(err))
		return false
	}
	logger.Debug("Request sent to controller")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

func isServiceInWorker(worker Worker, service string) (bool, string) {
	for _, v := range worker.Services {
		if v.Name == service {
			return true, v.Status
		}
	}
	return false, ""
}

func updateEditLastSopt(worker_id string, service string, lastSopt StartOptions) {
	worker := workers[worker_id]
	worker.lastSopt[service] = lastSopt
	workers[worker_id] = worker
}
