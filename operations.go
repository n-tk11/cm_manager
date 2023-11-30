package main

type Worker struct {
	ipAddr string
	port   string
	status string
}

type Service struct {
	name     string
	chkFiles []string
}

func checkpointService(worker Worker, service Service) {

}

func restoreService(worker Worker, service Service, checkpoint string) {

}

func migrateService(src Worker, dest Worker, service Service) {

}

func startService() {

}

func StopService() {

}
