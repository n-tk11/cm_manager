package main

import (
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func stopService(worker Worker, service Service) error {
	logger.Debug("Stopping service", zap.String("worker", worker.Id), zap.String("service", service.Name))
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/stop/" + service.Name

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Error("Error creating request", zap.Error(err))
		return err
	}
	req.Close = true
	client := &http.Client{}
	logger.Debug("Sending request to controller", zap.String("url", url))
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error sending request", zap.Error(err))
		return err
	}
	logger.Debug("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading the response body", zap.Error(err))
		return err
	}
	if resp.StatusCode != 200 {
		logger.Error("Stop service fail at worker", zap.String("worker", worker.Id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("stop service fail at worker with response code %d", resp.StatusCode)
	}
	deleteRunService(worker.Id, service.Name)
	logger.Info("Stop service at worker succesfully", zap.String("worker", worker.Id), zap.String("service", service.Name))
	return nil
}
