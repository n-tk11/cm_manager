package main

import "github.com/docker/docker/api/types/mount"

type Worker struct {
	IpAddrPort string //ex. 192.168.1.2:8787
	Status     string
}

type Service struct {
	Name     string
	ChkFiles []string
	Image    string
}

type StartBody struct {
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
