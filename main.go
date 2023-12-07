package main

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
)

func main() {
	var i int
	runopt := RunOptions{
		AppArgs:        "bash -c 'for i in $(seq 100); do echo $i > /checkpointfs/ff_out; sleep 1; done'",
		ImageURL:       "file:/checkpointfs/ff",
		OnAppReady:     "",
		PassphraseFile: "",
		PreservedPaths: "",
		NoRestore:      false,
		AllowBadImage:  false,
		LeaveStopped:   false,
		Verbose:        3,
		Envs:           []string{},
	}

	checkopt := CheckpointOptions{
		LeaveRun:      false,
		ImgUrl:        "",
		Passphrase:    "",
		Preserve_path: "",
		Num_shards:    4,
		Cpu_budget:    "",
		Verbose:       3,
		Envs:          []string{},
	}
	oneMount := mount.Mount{
		Type:   "bind",
		Source: "/mnt/checkpointfs",
		Target: "/checkpointfs",
	}
	secondMount := mount.Mount{
		Type:   "bind",
		Source: "/home/tul/MyProj/controller/services/ff",
		Target: "/opt/controller",
	}
	mounts := []mount.Mount{oneMount, secondMount}
	startopt := StartOptions{
		services["ff"].Name,
		services["ff"].Image,
		"80:80",
		[]string{},
		mounts,
		[]string{},
	}
	for {
		fmt.Println("Choose operations")
		fmt.Println("1 addService")
		fmt.Println("2 startService")
		fmt.Println("3 addWorker")
		fmt.Println("4 checkpointService")
		fmt.Println("5 runService")
		fmt.Println("6 checkpoint")
		fmt.Scan(&i)
		if i == 1 {
			fmt.Println("addService: name img")

			addService("ff", "ffdev:c4")
		} else if i == 2 {
			fmt.Println("startService")
			startServiceContainer(workers[0], startopt)
		} else if i == 3 {
			var w string
			fmt.Scan(&w)
			addWorker(w)
		} else if i == 4 {
			checkpointService(workers[0], services["ff"], checkopt)
		} else if i == 5 {
			runService(workers[0], services["ff"], runopt)
		} else if i == 6 {
			migrateService(workers[0], workers[1], services["ff"], checkopt, runopt, startopt, true)
		}
	}
}
