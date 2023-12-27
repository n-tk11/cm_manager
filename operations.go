package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var services = make(map[string]Service)

// var worker_count = 0
var workers = make(map[string]Worker)

func checkpointService(worker_id string, service Service, option CheckpointOptions) (string, error) {
	logger.Debug("Checkpointing service", zap.String("service", service.Name))
	url := "http://" + workers[worker_id].IpAddrPort + "/cm_controller/v1/checkpoint/" + service.Name
	currentTime := time.Now().UTC()

	// Format the time in ISO 8601 format
	iso8601Format := "2006-01-02T15:04:05Z07:00"
	iso8601Time := currentTime.Format(iso8601Format)
	option.ImgUrl = "file:/checkpointfs/" + service.Name + "_" + worker_id + "_" + iso8601Time
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
		logger.Info("Checkpoint successfully the image name", zap.String("image", option.ImgUrl))
		addCheckpointFile(service.Name, option.ImgUrl)
		return option.ImgUrl, nil
	} else {
		logger.Error("Checkpoint service fail at worker", zap.String("worker", worker_id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("checkpoint service fail at worker with response code %d", resp.StatusCode)

	}
}

// TODO add error handling
func migrateService(src string, dest string, service Service, copt CheckpointOptions, ropt RunOptions, sopt StartOptions, stopSrc bool) error {
	logger.Debug("Migrating service", zap.String("service", service.Name))
	migrateStart := time.Now()
	sErr := startServiceContainer(workers[dest], sopt)
	if sErr != nil {
		logger.Error("Error starting service's container at destination", zap.String("serviceName", service.Name), zap.String("dest", dest), zap.Error(sErr))
		return sErr
	}
	var cErr error
	ropt.ImageURL, cErr = checkpointService(src, service, copt)
	//Let user manage what port there want to use
	if cErr != nil {
		logger.Error("Error checkpoint service at source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(cErr))

		return cErr
	}
	//startServiceContainer(workers[dest], sopt)
	//time.Sleep(200 * time.Millisecond) //If too fast ffd may not ready
	rErr := runService(workers[dest], service, ropt)
	if rErr != nil {
		logger.Error("Failed to run service on destination, will start the service on source again", zap.String("serviceName", service.Name), zap.String("src", src), zap.String("dest", dest), zap.Error(rErr))

		rErr := runService(workers[src], service, ropt)
		if rErr != nil {
			logger.Error("Failed to rerun service on source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(rErr))
		}
		return rErr
	}
	if stopSrc {
		stErr := stopService(workers[src], service)
		if stErr != nil {
			logger.Error("Failed to stop service on source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(rErr))
			return stErr
		}
	}
	migrateEnd := time.Since(migrateStart)
	logger.Info("Migrate service successfully", zap.String("service", service.Name), zap.String("src", src), zap.String("dest", dest), zap.Duration("time", migrateEnd))

	return nil
}

func startServiceContainer(worker Worker, startBody StartOptions) error {
	logger.Debug("Starting service", zap.String("service", startBody.ContainerName))
	if _, ok := services[startBody.ContainerName]; ok {
		url := "http://" + worker.IpAddrPort + "/cm_controller/v1/start"
		reqJson := startBody

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

func addService(name string, image string) (Service, error) {
	newService := Service{
		Name:     name,
		ChkFiles: []string{},
		Image:    image,
	}
	if _, ok := services[name]; !ok {
		services[name] = newService
		logger.Debug("Service added", zap.String("serviceName", name))
		return newService, nil
	} else {
		logger.Error("Service already existed", zap.String("serviceName", name))
		return newService, errors.New("Service already existed")
	}
}

func addWorker(worker_id string, ipAddrPort string) (Worker, error) {
	newWorker := Worker{
		Id:         worker_id,
		IpAddrPort: ipAddrPort,
		Status:     "new",
		Services:   []ServiceInWorker{},
	}
	if _, ok := workers[worker_id]; !ok {
		workers[worker_id] = newWorker
		logger.Debug("Worker added", zap.String("workerID", worker_id))
		return newWorker, nil
	} else {
		logger.Error("Worker already existed", zap.String("WorkerId", worker_id))
		return newWorker, errors.New("Worker already existed")
	}
}

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
	logger.Info("Stop service at worker succesfully", zap.String("worker", worker.Id), zap.String("service", service.Name))
	return nil
}

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
		logger.Error("Run service fail at worker", zap.String("worker", worker.Id), zap.String("service", service.Name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("run service fail at worker with response code %d", resp.StatusCode)
	}

	logger.Info("Run service at worker succesfully", zap.String("worker", worker.Id), zap.String("service", service.Name))
	return nil
}

func addCheckpointFile(name string, path string) {
	if s, ok := services[name]; !ok {
		fmt.Printf("Service with name %s not found\n", name)
	} else {
		tmp := s
		tmp.ChkFiles = append(tmp.ChkFiles, path)
		services[name] = tmp
	}

}

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

func deleteService(name string) {
	delete(services, name)
}
