package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func removeService(worker Worker, service Service) error {
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/remove/" + service.Name
	logger.Debug("Removing service", zap.String("worker", worker.Id), zap.String("service", service.Name))

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(nil))
	if err != nil {
		logger.Error("Error creating request", zap.Error(err))
		return err
	}
	req.Close = true
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
		logger.Error("Remove service fail at worker", zap.String("worker", worker.Id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("remove service fail at worker with response code %d", resp.StatusCode)
	}
	deleteService(service.Name)
	logger.Info("Remove service at worker succesfully", zap.String("worker", worker.Id), zap.String("service", service.Name))
	return nil
}
