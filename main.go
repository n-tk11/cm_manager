package main

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
)

func main() {
	var i int
	for {
		fmt.Println("Choose operations")
		fmt.Scan(&i)
		if i == 1 {
			fmt.Println("addService: name img")

			addService("ff", "ffdev:c4")
		} else if i == 2 {
			fmt.Println("startService")
			appPort := "80:80"

			oneMount := mount.Mount{
				Type:   "bind",
				Source: "/tmp",
				Target: "/tmp",
			}
			secondMount := mount.Mount{
				Type:   "bind",
				Source: "/home/tul/MyProj/controller/services/ff",
				Target: "/opt/controller",
			}
			mounts := []mount.Mount{oneMount, secondMount}
			envs := []string{}
			caps := []string{}
			startServiceContainer(workers[0], "ff", appPort, envs, mounts, caps)
		} else if i == 3 {
			addWorker("127.0.0.1:8787")
		} else if i == 4 {
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
			checkpointService(workers[0], services["ff"], checkopt)
		} else if i == 5 {
			runopt := RunOptions{
				AppArgs:        "bash -c 'for i in $(seq 100); do echo $i > /tmp/ff_out; sleep 1; done",
				ImageURL:       "file:/tmp/ff",
				OnAppReady:     "",
				PassphraseFile: "",
				PreservedPaths: "",
				NoRestore:      false,
				AllowBadImage:  false,
				LeaveStopped:   false,
				Verbose:        3,
				Envs:           []string{},
			}
			runService(workers[0], services["ff"], runopt)
		}
	}
}
