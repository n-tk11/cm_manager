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
