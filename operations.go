package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var services = make(map[string]Service)
var workers = []Worker{}

func checkpointService(worker_id int, service Service, option CheckpointOptions) string {

	url := "http://" + workers[worker_id].IpAddrPort + "/cm_checkpoint/" + service.Name
	currentTime := time.Now().UTC()

	// Format the time in ISO 8601 format
	iso8601Format := "2006-01-02T15:04:05Z07:00"
	iso8601Time := currentTime.Format(iso8601Format)
	option.ImgUrl = "file:/checkpointfs/" + service.Name + "_" + strconv.Itoa(worker_id) + "_" + iso8601Time
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
		return option.ImgUrl
	}
	fmt.Println("Checkpoint failed")
	return ""
}

func migrateService(src int, dest int, service Service, copt CheckpointOptions, ropt RunOptions, sopt StartOptions, stopSrc bool) {

	ropt.ImageURL = checkpointService(src, service, copt)
	//Let user manage what port there want to use
	if ropt.ImageURL == "" {
		fmt.Println("migrate failed")
		return
	}
	startServiceContainer(workers[dest], sopt)
	runService(workers[dest], service, ropt)
	if stopSrc {
		stopService(workers[src], service)
	}
}

func startServiceContainer(worker Worker, startBody StartOptions) {
	fmt.Printf("Starting service %s\n", startBody.ContainerName)
	if _, ok := services[startBody.ContainerName]; ok {
		url := "http://" + worker.IpAddrPort + "/cm_start"
		reqJson := startBody

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

// TODO TEST
func stopService(worker Worker, service Service) {
	url := "http://" + worker.IpAddrPort + "/cm_stop/" + service.Name

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating new request:", err)
		return
	}
	req.Close = true
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
