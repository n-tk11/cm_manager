package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var services = make(map[string]Service)
var worker_count = 0
var workers = make(map[string]Worker)

func checkpointService(worker_id string, service Service, option CheckpointOptions) string {

	url := "http://" + workers[worker_id].IpAddrPort + "/cm_controller/v1/checkpoint/" + service.Name
	currentTime := time.Now().UTC()

	// Format the time in ISO 8601 format
	iso8601Format := "2006-01-02T15:04:05Z07:00"
	iso8601Time := currentTime.Format(iso8601Format)
	option.ImgUrl = "file:/checkpointfs/" + service.Name + "_" + worker_id + "_" + iso8601Time
	requestBody, err := json.Marshal(option)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return ""
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request")
		return ""
	}
	fmt.Println("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the responseBody")
		return ""
	}
	fmt.Printf("%d\n %s\n", resp.StatusCode, string(body))
	if resp.StatusCode == 200 {
		fmt.Printf("Checkpoint successfully the image name %s\n", option.ImgUrl)
		addCheckpointFile(service.Name, option.ImgUrl)
		return option.ImgUrl
	}
	fmt.Println("Checkpoint failed")
	return ""
}

func migrateService(src string, dest string, service Service, copt CheckpointOptions, ropt RunOptions, sopt StartOptions, stopSrc bool) int {

	sErr := startServiceContainer(workers[dest], sopt)
	if sErr != 0 {
		fmt.Println("Failed to start service on destination")
		return 1
	}
	ropt.ImageURL = checkpointService(src, service, copt)
	//Let user manage what port there want to use
	if ropt.ImageURL == "" {
		fmt.Println("Fail to checkpoint service on source")
		return 1
	}
	//startServiceContainer(workers[dest], sopt)
	//time.Sleep(200 * time.Millisecond) //If too fast ffd may not ready
	rErr := runService(workers[dest], service, ropt)
	if rErr != 0 {
		fmt.Println("Failed to run service on destination, with start the service on source again")
		rErr := runService(workers[src], service, ropt)
		if rErr != 0 {
			fmt.Println("Failed to rerun service on source")
		}
		return 1
	}
	if stopSrc {
		stErr, _ := stopService(workers[src], service)
		if stErr != 200 {
			fmt.Println("Failed to stop service on source")
			return 1
		}
	}
	return 0
}

func startServiceContainer(worker Worker, startBody StartOptions) int {
	fmt.Printf("Starting service %s\n", startBody.ContainerName)
	if _, ok := services[startBody.ContainerName]; ok {
		url := "http://" + worker.IpAddrPort + "/cm_controller/v1/start"
		reqJson := startBody

		requestBody, err := json.Marshal(reqJson)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return 1
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		req.Close = true
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending the request")
			return 1
		}
		fmt.Println("Request sent to controller")
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the responseBody")
			return 1
		}
		fmt.Printf("%d %s\n", resp.StatusCode, string(body))
		if resp.StatusCode == 200 {
			fmt.Printf("Service %s started\n", startBody.ContainerName)
			return 0
		} else {
			fmt.Printf("Service %s start failed\n", startBody.ContainerName)
			return 1
		}
	} else {
		fmt.Println("Service not found, add the service first")
		return 1
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

func addWorker(worker_id string, ipAddrPort string) (Worker, int) {
	newWorker := Worker{
		IpAddrPort: ipAddrPort,
		Status:     "new",
	}
	if _, ok := workers[worker_id]; !ok {
		workers[worker_id] = newWorker
		return newWorker, 0
	} else {
		fmt.Printf("Worker with id %s already existed\n", worker_id)
		return newWorker, 1
	}
}

// TODO TEST
func stopService(worker Worker, service Service) (int, string) {
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/stop/" + service.Name

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating new request:", err)
		return 1, "Error creating new request"
	}
	req.Close = true
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request")
		return 1, "Error sending the request"
	}
	fmt.Println("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the responseBody")
		return 1, "Error reading the responseBody"
	}
	fmt.Printf("%d %s\n", resp.StatusCode, string(body))
	return resp.StatusCode, string(body)
}

func runService(worker Worker, service Service, option RunOptions) int {
	url := "http://" + worker.IpAddrPort + "/cm_controller/v1/run/" + service.Name

	requestBody, err := json.Marshal(option)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return 1
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request")
		return 1
	}
	fmt.Println("Request sent to controller")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the responseBody")
		return 1
	}
	fmt.Printf("%d %s\n", resp.StatusCode, string(body))
	return 0
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
