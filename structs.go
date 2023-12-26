package main

import "github.com/docker/docker/api/types/mount"

type Worker struct {
	Id         string            `json:"id"`
	IpAddrPort string            `json:"addr"` //ex. 192.168.1.2:8787
	Status     string            `json:"status"`
	Services   []ServiceInWorker `json:"services"`
}

type ServiceInWorker struct {
	Name   string `json:"name"`
	Stutus string `json:"status"`
}

type Service struct {
	Name     string   `json:"name"`
	ChkFiles []string `json:"chk_files"`
	Image    string   `json:"image"`
}

type StartOptions struct {
	ContainerName string        `json:"container_name"`
	Image         string        `json:"image"`
	AppPort       string        `json:"app_port"`
	Envs          []string      `json:"envs"`
	Mounts        []mount.Mount `json:"mounts"`
	Caps          []string      `json:"caps"`
}

type CheckpointOptions struct {
	LeaveRun      bool     `json:"leave_running"`
	ImgUrl        string   `json:"image_url"`
	Passphrase    string   `json:"passphrase_file"`
	Preserve_path string   `json:"preserved_paths"`
	Num_shards    int      `json:"num_shards"`
	Cpu_budget    string   `json:"cpu_budget"`
	Verbose       int      `json:"verbose"`
	Envs          []string `json:"envs"`
}

type RunOptions struct {
	AppArgs        string   `json:"app_args"`
	ImageURL       string   `json:"image_url"`
	OnAppReady     string   `json:"on_app_ready"`
	PassphraseFile string   `json:"passphrase_file"`
	PreservedPaths string   `json:"preserved_paths"`
	NoRestore      bool     `json:"no_restore"`
	AllowBadImage  bool     `json:"allow_bad_image"`
	LeaveStopped   bool     `json:"leave_stopped"`
	Verbose        int      `json:"verbose"`
	Envs           []string `json:"envs"`
}

func addRunService(workerId string, service ServiceInWorker) {
	worker := workers[workerId]
	worker.Services = append(worker.Services, service)
	workers[workerId] = worker
}

func deleteRunService(workerId string, service string) {
	worker := workers[workerId]
	for i, v := range worker.Services {
		if v.Name == service {
			worker.Services = append(worker.Services[:i], worker.Services[i+1:]...)
		}
	}
	workers[workerId] = worker
}

func updateRunService(workerId string, service ServiceInWorker) {
	worker := workers[workerId]
	for i, v := range worker.Services {
		if v.Name == service.Name {
			worker.Services[i] = service
		}
	}
	workers[workerId] = worker
}
