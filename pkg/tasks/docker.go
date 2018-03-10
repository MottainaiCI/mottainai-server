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
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
)

func DockerExecute(docID string) (int, error) {
	fetcher := client.NewFetcher(docID)
	th := DefaultTaskHandler()

	task_info := th.FetchTask(fetcher)
	if task_info.Status == "running" {
		fetcher.SetTaskStatus("failure")
		msg := "Task picked twice"
		fetcher.AppendTaskOutput(msg)
		return 1, errors.New(msg)
	}

	fetcher.SetTaskStatus("running")
	fetcher.SetTaskField("start_time", time.Now().Format("20060102150405"))
	fetcher.AppendTaskOutput("Build started!\n")
	task_info = th.FetchTask(fetcher)

	dir, err := ioutil.TempDir(setting.Configuration.TempWorkDir, docID)
	if err != nil {
		panic(err)
	}

	artdir, err := ioutil.TempDir(setting.Configuration.TempWorkDir, "artefact")
	if err != nil {
		panic(err)
	}

	storagetmp, err := ioutil.TempDir(setting.Configuration.TempWorkDir, "storage")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(artdir)
	defer os.RemoveAll(storagetmp)
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
	if len(task_info.Commit) > 0 {
		out, err = utils.Git([]string{"checkout", task_info.Commit}, git_repo_dir)
		fetcher.AppendTaskOutput(out)
		if err != nil {
			panic(err)
		}
	}

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
	var git_root_path = path.Join(setting.Configuration.BuildPath, strconv.Itoa(task_info.ID))
	defer os.RemoveAll(git_root_path)
	var git_build_root_path = path.Join(git_root_path, task_info.Directory)

	var storage_path = "storage"
	var artefact_path = "artefacts"

	if len(task_info.ArtefactPath) > 0 {
		artefact_path = task_info.ArtefactPath
	}

	if len(task_info.StoragePath) > 0 {
		storage_path = task_info.StoragePath
	}

	var storage_root_path = path.Join(git_build_root_path, storage_path)

	var ContainerBinds []string

	var artefactdir string
	var storagedir string

	if setting.Configuration.DockerInDocker {
		ContainerBinds = append(ContainerBinds, setting.Configuration.DockerEndpointDiD+":/var/run/docker.sock")
		ContainerBinds = append(ContainerBinds, "/tmp:/tmp")
		ContainerBinds = append(ContainerBinds, path.Join(git_build_root_path, artefact_path)+":"+path.Join(git_build_root_path, artefact_path))
		ContainerBinds = append(ContainerBinds, storage_root_path+":"+storage_root_path)

		artefactdir = path.Join(git_build_root_path, artefact_path)
		storagedir = storage_root_path
	} else {
		ContainerBinds = append(ContainerBinds, artdir+":"+path.Join(git_build_root_path, artefact_path))
		ContainerBinds = append(ContainerBinds, storagetmp+":"+storage_root_path)

		artefactdir = artdir
		storagedir = storagetmp
	}

	if len(task_info.RootTask) > 0 {
		fetcher.DownloadArtefactsFromTask(task_info.RootTask, artefactdir)
	}

	if len(task_info.Namespace) > 0 {
		fetcher.DownloadArtefactsFromNamespace(task_info.Namespace, artefactdir)
	}

	if len(task_info.Storage) > 0 {
		fetcher.DownloadArtefactsFromStorage(task_info.Storage, storagedir)
	}

	//ContainerVolumes = append(ContainerVolumes, git_repo_dir+":/build")

	ContainerBinds = append(ContainerBinds, git_repo_dir+":"+git_root_path)

	createContHostConfig := docker.HostConfig{
		Privileged: setting.Configuration.DockerPriviledged,
		Binds:      ContainerBinds,
		//	LogConfig:  docker.LogConfig{Type: "json-file"}
	}

	var containerconfig = &docker.Config{
		Image: task_info.Image,
		Cmd:   []string{"-c", "pwd;ls -liah;" + execute_script},
		//	Env:        config.Env,
		WorkingDir: git_build_root_path,
		Entrypoint: []string{"/bin/sh"},
		//Entrypoint:  //[]string{execute_script},
	}

	fetcher.AppendTaskOutput("Binds: ")
	for _, v := range ContainerBinds {
		fetcher.AppendTaskOutput("- " + v)
	}

	fetcher.AppendTaskOutput("Container working dir: " + git_build_root_path)

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
	defer CleanUpContainer(docker_client, container.ID)
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
		task_info = th.FetchTask(fetcher)
		if task_info.Status != "running" {
			fetcher.AppendTaskOutput("Aborting execution")
			docker_client.StopContainer(container.ID, uint(20))
			fetcher.SetTaskResult("stopped")
			fetcher.SetTaskStatus("stop")
			return 0, nil
		}
		c_data, err := docker_client.InspectContainer(container.ID) // update our container information
		if err != nil {
			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			fetcher.AppendTaskOutput(err.Error())
			return 0, nil
		}
		if c_data.State.Running == false {

			var err error

			to_upload := artdir
			if setting.Configuration.DockerInDocker {
				to_upload = path.Join(git_root_path, task_info.Directory, artefact_path)
			}

			err = filepath.Walk(to_upload, func(path string, f os.FileInfo, err error) error {
				return th.UploadArtefact(fetcher, path, to_upload)

			})

			if err != nil {
				fetcher.AppendTaskOutput(err.Error())
			}

			fetcher.AppendTaskOutput("Container execution terminated")
			return c_data.State.ExitCode, nil
		}
	}

}

func CleanUpContainer(client *docker.Client, ID string) {
	client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ID,
		Force: true,
	})
}
