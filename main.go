package main

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
)

func main() {
	var i int
	runopt := RunOptions{
		AppArgs:        "bash -c 'for i in $(seq 100); do echo $i > /tmp/ff_out; sleep 1; done'",
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
	secdMount := mount.Mount{
		Type:   "bind",
		Source: "/tmp",
		Target: "/tmp",
	}
	mounts := []mount.Mount{oneMount, secdMount}
	startopt := StartOptions{
		ContainerName: services["ff"].Name,
		Image:         services["ff"].Image,
		AppPort:       "80:80",
		Envs:          []string{},
		Mounts:        mounts,
		Caps:          []string{},
	}
	for {
		fmt.Println("Choose operations")
		fmt.Println("1 addService")
		fmt.Println("2 startService")
		fmt.Println("3 addWorker")
		fmt.Println("4 checkpointService")
		fmt.Println("5 runService")
		fmt.Println("6 migrate")
		fmt.Scan(&i)
		if i == 1 {
			fmt.Println("addService: name img")

			addService("ff", "ffdev:c4")
			startopt.ContainerName = services["ff"].Name
			startopt.Image = services["ff"].Image
		} else if i == 2 {
			var x int
			fmt.Scan(&x)
			fmt.Println("startService")
			startServiceContainer(workers[x], startopt)
		} else if i == 3 {
			var w string
			fmt.Scan(&w)
			addWorker(w)
		} else if i == 4 {
			var x int
			fmt.Scan(&x)
			checkpointService(x, services["ff"], checkopt)
		} else if i == 5 {
			var x int
			fmt.Scan(&x)
			runService(workers[x], services["ff"], runopt)
		} else if i == 6 {
			startopt.Mounts[0].Type = "volume"
			startopt.Mounts[0].Source = "chkfs"
			startopt.Mounts[0].Target = "/checkpointfs"
			var x, y int
			fmt.Scan(&x)
			fmt.Scan(&y)
			migrateService(x, y, services["ff"], checkopt, runopt, startopt, true)

		}
	}
}
