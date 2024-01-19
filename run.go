package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

var runCount = 0

func runService(worker Worker, service Service, option RunOptions) error {
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/run/" + service.Name
	logger.Debug("Running service", zap.String("worker", worker.Id), zap.String("service", service.Name))
	requestBody, err := json.Marshal(option)
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
		return err
	}
	logger.Debug("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading the responseBody", zap.Error(err))
		return err
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 500 && runCount < 1 {
			runCount++
			logger.Error("Run Error 500 will try again", zap.String("worker", worker.Id), zap.String("service", service.Name))
			return runService(worker, service, option)

		}
		logger.Error("Run service fail at worker", zap.String("worker", worker.Id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("run service fail at worker with response code %d", resp.StatusCode)
	}
	config := serviceConfigs[service.Name]
	config.RunOpt = option
	serviceConfigs[service.Name] = config

	logger.Info("Run service at worker succesfully", zap.String("worker", worker.Id), zap.String("service", service.Name))
	return nil
}
