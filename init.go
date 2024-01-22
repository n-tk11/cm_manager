package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

func manager_init() {
	logger := getGlobalLogger()
	logger.Debug("Initializing manager")
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--workers" || args[i] == "-w" {
			workerPath := args[i+1]
			worker_init(workerPath)
		}
		if args[i] == "--services" || args[i] == "-s" {
			servicePath := args[i+1]
			service_init(servicePath)
		}
	}
	scanServicesOnWorkers()
	scanCheckpointFiles(0, "")
}

func worker_init(workerPath string) {
	file, err := os.Open(workerPath)
	if err != nil {
		logger.Error("Error opening WorkerFile", zap.Error(err))
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Iterate over each line
	for scanner.Scan() {
		line := scanner.Text()
		// Process each line
		if strings.HasPrefix(line, "#") {
			continue
		}
		worker_id, addr := processLine(line)
		if worker_id == "" || addr == "" {
			continue
		}
		addWorker(worker_id, addr)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		logger.Error("Error reading WorkerFile", zap.Error(err))
	}
}

func service_init(servicePath string) {
	file, err := os.Open(servicePath)
	if err != nil {
		logger.Error("Error opening ServiceFile", zap.Error(err))
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Iterate over each line
	for scanner.Scan() {
		line := scanner.Text()
		// Process each line
		if strings.HasPrefix(line, "#") {
			continue
		}
		name, image := processLine(line)
		if name == "" || image == "" {
			continue
		}
		addService(name, image)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		logger.Error("Error reading ServiceFile", zap.Error(err))
	}
}

func scanServicesOnWorkers() {
	for _, worker := range workers {
		worker_id := worker.Id
		for _, v := range services {
			status, err := queryServiceStatus(worker_id, v.Name)
			if err != nil {
				continue
			}
			if status != "" {
				logger.Debug("Adding service to a worker(run)", zap.String("worker_id", worker_id), zap.String("service_name", v.Name), zap.String("status", status))
				addRunService(worker_id, ServiceInWorker{Name: v.Name, Status: status})
			}

		}
	}
}

func processLine(line string) (string, string) {
	// Split the line into two values
	values := strings.Fields(line)

	// Check if there are at least two values in the line
	if len(values) >= 2 {
		// Assuming 'a' and 'b' are integers in this example
		a := values[0]
		b := values[1]

		// Perform some action with 'a' and 'b'

		return a, b
		// Add your custom logic here
	} else {
		fmt.Println("Invalid line format:", line)
		return "", ""
	}
}

func scanCheckpointFiles(mode int, service string) {
	//mode = 0 -> scan all services
	//mode = 1 -> scan specific service
	logger.Debug("Checking services")
	dirPath := "/mnt/checkpointfs/"
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("Error reading services dir", zap.Error(err))
		return
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		fileName := dirEntry.Name()
		fields := strings.Split(fileName, "_")
		if len(fields) >= 1 {
			serviceName := fields[0]
			if mode == 1 && serviceName != service {
				continue
			}
			if _, ok := services[serviceName]; ok {
				addCheckpointFile(serviceName, "file:/checkpointfs/"+fileName)
			}
		}
	}
}
