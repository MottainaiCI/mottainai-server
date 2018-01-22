/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package agenttasks

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
)

func Execute(docID string) (int, error) {
	fetcher := client.NewFetcher(docID)
	fetcher.SetTaskStatus("running")
	fetcher.SetTaskOutput("Build started!\n")

	task_info := FetchTask(fetcher, docID)
	dir, err := ioutil.TempDir(setting.Configuration.TempWorkDir, task_info.Namespace)
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir)

	fetcher.AppendTaskOutput("Cloning git repo: " + task_info.Source)

	out, err := utils.Git([]string{"clone", task_info.Source, "target_repo"}, dir)
	fetcher.AppendTaskOutput(out)
	if err != nil {
		panic(err)
	}

	git_repo_dir := filepath.Join(dir, "target_repo")

	//cwd, _ := os.Getwd()
	os.Chdir(git_repo_dir)

	var execute_script = "mottainai-run"

	if len(task_info.Script) > 0 {
		execute_script = task_info.Script
	}
	// XXX: To replace with PID handling and background process.
	// XXX: Exp. in docker container
	// XXX: Start with docker, monitor it.
	// XXX: optional args, with --priviledged and -v socket
	docker_client, err := docker.NewClient(setting.Configuration.DockerEndpoint)
	if err != nil {
		panic(errors.New(err.Error() + " ENDPOINT:" + setting.Configuration.DockerEndpoint))
	}

	if len(task_info.Image) > 0 {
		fetcher.AppendTaskOutput("Pulling image: " + task_info.Image)
		if err = docker_client.PullImage(docker.PullImageOptions{Repository: task_info.Image}, docker.AuthConfiguration{}); err != nil {
			panic(err)
		}
		fetcher.AppendTaskOutput("Pulling image: DONE!")
	}
	//var args []string

	var ContainerBinds []string

	if setting.Configuration.DockerInDocker {
		ContainerBinds = append(ContainerBinds, setting.Configuration.DockerEndpointDiD+":/var/run/docker.sock")
		ContainerBinds = append(ContainerBinds, "/tmp:/tmp")
	}
	//ContainerVolumes = append(ContainerVolumes, git_repo_dir+":/build")
	ContainerBinds = append(ContainerBinds, git_repo_dir+":/build")

	createContHostConfig := docker.HostConfig{
		Privileged: setting.Configuration.DockerPriviledged,
		Binds:      ContainerBinds,
		//	LogConfig:  docker.LogConfig{Type: "json-file"}
	}

	var containerconfig = &docker.Config{
		Image: task_info.Image,
		Cmd:   []string{"-c", "ls -liah;" + execute_script},
		//	Env:        config.Env,
		WorkingDir: "/build" + task_info.Directory,
		Entrypoint: []string{"/bin/sh"},
		//Entrypoint:  //[]string{execute_script},
	}

	fetcher.AppendTaskOutput("Binds: ")
	for _, v := range ContainerBinds {
		fetcher.AppendTaskOutput("- " + v)
	}

	fetcher.AppendTaskOutput("Container working dir: " + "/build/" + task_info.Directory)

	container, err := docker_client.CreateContainer(docker.CreateContainerOptions{
		Config:     containerconfig,
		HostConfig: &createContHostConfig,
	})

	if err != nil {
		panic(err)
	}

	utils.ContainerOutputAttach(func(s string) {
		fetcher.AppendTaskOutput(s)
	}, docker_client, container)
	//defer CleanUpContainer(docker_client, container.ID)
	if setting.Configuration.DockerKeepImg == false {
		defer docker_client.RemoveImage(task_info.Image)
	}

	fetcher.AppendTaskOutput("Created container ID: " + container.ID)

	err = docker_client.StartContainer(container.ID, &createContHostConfig)
	if err != nil {
		panic(err)
	}
	fetcher.AppendTaskOutput("Started Container " + container.ID)

	for {
		time.Sleep(1 * time.Second)
		task_info = FetchTask(fetcher, docID)
		if task_info.Status == "stop" {
			fetcher.AppendTaskOutput("Asked to stop")
			docker_client.StopContainer(container.ID, uint(20))
			fetcher.SetTaskResult("stopped")
			fetcher.SetTaskStatus("stop")
			return 0, nil
		}
		c_data, err := docker_client.InspectContainer(container.ID) // update our container information
		if err != nil {
			panic(err)
		}
		if c_data.State.Running == false {
			fetcher.AppendTaskOutput("Container execution terminated")
			return c_data.State.ExitCode, nil
		}
	}

	//fetcher := client.NewFetcher()
	//SetTaskStatus(docID, "done")
	//panic(errors.New("oops"))
	fetcher.SetTaskResult("error")
	fetcher.SetTaskStatus("done")
	return 0, nil
}

func CleanUpContainer(client *docker.Client, ID string) {
	client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ID,
		Force: true,
	})
}

func HandleSuccess(result int, docID string) error {
	fetcher := client.NewFetcher(docID)

	fetcher.SetTaskField("exit_status", strconv.Itoa(result))
	fetcher.SetTaskResult("success")
	fetcher.SetTaskStatus("done")
	return nil
}

func HandleErr(errstring, docID string) error {
	fetcher := client.NewFetcher(docID)

	fetcher.AppendTaskOutput(errstring)
	fetcher.SetTaskResult("error")
	fetcher.SetTaskStatus("done")
	return nil
}
