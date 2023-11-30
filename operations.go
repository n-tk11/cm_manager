package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/mount"
)

var services = make(map[string]Service)
var workers = []Worker{}

func checkpointService(worker Worker, service Service, option CheckpointOptions) {

	url := "http://" + worker.IpAddrPort + "/cm_checkpoint/" + service.Name

	requestBody, err := json.Marshal(option)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request")
		return
	}
	fmt.Println("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the responseBody")
		return
	}
	fmt.Printf("%d\n %s\n", resp.StatusCode, string(body))
}

func restoreService(worker Worker, service Service, checkpoint string) {

}

func migrateService(src Worker, dest Worker, service Service) {

}

func startServiceContainer(worker Worker, serviceName string, appPort string, envs []string, mounts []mount.Mount, caps []string) {
	if service, ok := services[serviceName]; ok {
		url := "http://" + worker.IpAddrPort + "/cm_start"
		reqJson := StartBody{
			ContainerName: service.Name,
			Image:         service.Image,
			AppPort:       appPort,
			Envs:          envs,
			Mounts:        mounts,
			Caps:          caps,
		}

		requestBody, err := json.Marshal(reqJson)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		req.Close = true
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending the request")
			return
		}
		fmt.Println("Request sent to controller")
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the responseBody")
			return
		}
		fmt.Printf("%d %s\n", resp.StatusCode, string(body))
	} else {
		fmt.Println("Service not found, add the service first")
	}
}

func addService(name string, image string) (Service, int) {
	newService := Service{
		Name:     name,
		ChkFiles: []string{},
		Image:    image,
	}
	if _, ok := services[name]; !ok {
		services[name] = newService
		return newService, 0
	} else {
		fmt.Printf("Service with name %s already existed\n", name)
		return newService, 1
	}
}

func addWorker(ipAddrPort string) Worker {
	newWorker := Worker{
		IpAddrPort: ipAddrPort,
		Status:     "new",
	}
	workers = append(workers, newWorker)
	return newWorker
}

func StopService() {

}

func runService(worker Worker, service Service, option RunOptions) {
	url := "http://" + worker.IpAddrPort + "/cm_run/" + service.Name

	requestBody, err := json.Marshal(option)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request")
		return
	}
	fmt.Println("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the responseBody")
		return
	}
	fmt.Printf("%d %s\n", resp.StatusCode, string(body))
}
