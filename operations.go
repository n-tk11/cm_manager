package main

import (
	"errors"
	"fmt"

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
		services[name] = newService
		serviceConfigs[name] = newServiceConfig
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

func addCheckpointFile(name string, path string) {
	if s, ok := services[name]; !ok {
		fmt.Printf("Service with name %s not found\n", name)
	} else {
		tmp := s
		tmp.ChkFiles = append(tmp.ChkFiles, path)
		services[name] = tmp
	}

}

func deleteService(name string) {
	delete(services, name)
}

// Much operation functions now move into new seperate files(eg. start, run,remove,checkpoint,etc.)
