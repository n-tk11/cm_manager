package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"go.uber.org/zap"
)

var services = make(map[string]Service)
var serviceConfigs = make(map[string]ServiceConfig)

// var worker_count = 0
var workers = make(map[string]Worker)

func addService(name string, image string) (Service, error) {
	newService := Service{
		Name:     name,
		ChkFiles: []string{},
		Image:    image,
	}
	newServiceConfig := ServiceConfig{
		StartOpt: StartOptions{
			ContainerName: name,
			Image:         image,
			AppPort:       "",
			Envs:          []string{},
			Mounts:        []mount.Mount{},
			Caps:          []string{},
		},
		RunOpt: RunOptions{
			AppArgs:        "",
			ImageURL:       "",
			OnAppReady:     "",
			PassphraseFile: "",
			PreservedPaths: "",
			NoRestore:      false,
			AllowBadImage:  false,
			LeaveStopped:   false,
			Verbose:        0,
			Envs:           []string{},
		},
		ChkOpt: CheckpointOptions{
			LeaveRun:      false,
			ImgUrl:        "",
			Passphrase:    "",
			Preserve_path: "",
			Num_shards:    4,
			Cpu_budget:    "medium",
			Verbose:       0,
			Envs:          []string{},
		},
	}
	if _, ok := services[name]; !ok {
		err := mkChkDir(name)
		if err != nil {
			logger.Error("Error creating checkpoint directory", zap.Error(err))
			return newService, err
		}
		services[name] = newService
		serviceConfigs[name] = newServiceConfig

		logger.Debug("Service added", zap.String("serviceName", name))
		return newService, nil
	} else {
		logger.Error("Service already existed", zap.String("serviceName", name))
		return newService, errors.New("Service already existed")
	}
}

func addWorker(worker_id string, ipAddrPort string, init bool) (Worker, error) {
	newWorker := Worker{
		Id:         worker_id,
		IpAddrPort: ipAddrPort,
		Status:     "new",
		Services:   []ServiceInWorker{},
		countDown:  0,
		lastSopt:   make(map[string]StartOptions),
	}
	if _, ok := workers[worker_id]; !ok {
		workers[worker_id] = newWorker
		if !init {
			scanServicesOnAWorker(worker_id)
		}
		logger.Debug("Worker added", zap.String("workerID", worker_id))
		return newWorker, nil
	} else {
		logger.Error("Worker already existed", zap.String("WorkerId", worker_id))
		return newWorker, errors.New("Worker already existed")
	}
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

func deleteService(name string) error {
	for _, worker := range workers {
		for _, service := range worker.Services {
			if service.Name == name {
				status, err := queryServiceStatus(worker.Id, name)
				if err != nil {
					logger.Debug("Error querying service status", zap.Error(err))
					break
				} else if status == "running" || status == "paused" || status == "standby" || status == "checkpointed" {
					err := stopService(worker, services[name])
					if err != nil {
						logger.Error("Error stopping service", zap.Error(err))
						return err
					}
				}
				err = removeService(worker, services[name])
				if err != nil {
					logger.Error("Error removing service", zap.Error(err))
					return err
				}
				break
			}
		}
	}
	delete(services, name)
	return nil
}

func deleteServiceCheckpoint(name string) error {
	dirPath := "/mnt/checkpointfs/"
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("Error reading services dir", zap.Error(err))
		return err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		fileName := dirEntry.Name()
		fields := strings.Split(fileName, "_")
		if len(fields) >= 1 {
			serviceName := fields[0]
			if serviceName == name {
				err := os.RemoveAll(dirPath + fileName)
				if err != nil {
					logger.Error("Error removing service checkpoint", zap.Error(err))
					return err
				}
			}
		}
	}
	return nil
}

func unsubscribeService(worker_id string, name string) error {
	worker, ok := workers[worker_id]
	if !ok {
		fmt.Printf("Worker with id %s not found\n", worker_id)
		return errors.New("Worker not found")
	}
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/unsubscribe/" + name

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
		logger.Error("Unsubscribe Service Fail at worker", zap.String("worker", worker.Id), zap.String("service", name), zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("Unsubscribe Service Fail at workerr with response code %d", resp.StatusCode)
	}

	return nil
}

func deleteWorker(worker_id string) {
	delete(workers, worker_id)
}

func deleteCheckpointFiles(service string) error {
	logger.Debug("Deleting Chekpoint Files of a service ", zap.String("service", service))
	dirPath := "/mnt/checkpointfs/"
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("Error reading services dir", zap.Error(err))
		return err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		fileName := dirEntry.Name()
		fields := strings.Split(fileName, "_")
		if len(fields) >= 1 {
			serviceName := fields[0]
			if serviceName == service {
				logger.Debug("Deleting Chekpoint File", zap.String("file", dirPath+fileName))
				err := os.RemoveAll(dirPath + fileName)
				if err != nil {
					logger.Error("Error removing service checkpoint", zap.Error(err))
					return err
				}
			}
		}
	}
	return nil
}

func mkChkDir(name string) error {
	dirPath := "/mnt/checkpointfs/" + name
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logger.Debug("Creating checkpoint directory for service", zap.String("serviceName", name))
		err := os.Mkdir(dirPath, 0775)
		if err != nil {
			logger.Error("Error creating checkpoint directory", zap.Error(err))
			return err
		}
	}
	return nil
}

// Much operation functions now move into new seperate files(eg. start, run,remove,checkpoint,etc.)
