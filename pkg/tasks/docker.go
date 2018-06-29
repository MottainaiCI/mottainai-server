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
	"net/url"
	"path"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerExecutor struct {
	*TaskExecutor
	DockerClient *docker.Client
}

func NewDockerExecutor() *DockerExecutor {
	return &DockerExecutor{TaskExecutor: &TaskExecutor{Context: NewExecutorContext()}}
}

func (e *DockerExecutor) Prune() {
	e.DockerClient.PruneContainers(docker.PruneContainersOptions{})
	e.DockerClient.PruneImages(docker.PruneImagesOptions{})
	e.DockerClient.PruneVolumes(docker.PruneVolumesOptions{})
	e.DockerClient.PruneNetworks(docker.PruneNetworksOptions{})
}

func (d *DockerExecutor) Setup(docID string) error {
	d.TaskExecutor.Setup(docID)
	docker_client, err := docker.NewClient(setting.Configuration.DockerEndpoint)
	if err != nil {
		return (errors.New("Endpoint:" + setting.Configuration.DockerEndpoint + " Error: " + err.Error()))
	}
	d.DockerClient = docker_client
	return nil
}

func (d *DockerExecutor) Play(docID string) (int, error) {
	fetcher := d.MottainaiClient
	th := DefaultTaskHandler()
	task_info := th.FetchTask(fetcher)

	var sharedName, OriginalSharedName string
	image := task_info.Image

	u, err := url.Parse(task_info.Source)
	if err != nil {
		OriginalSharedName = image + task_info.Directory
	} else {
		OriginalSharedName = image + u.Path + task_info.Directory
	}

	sharedName, err = utils.StrictStrip(OriginalSharedName)
	if err != nil {
		panic(err)
	}

	artdir := d.Context.ArtefactDir
	storagetmp := d.Context.StorageDir
	git_repo_dir := d.Context.SourceDir

	var execute_script = "mottainai-run"

	if len(task_info.Script) > 0 {
		execute_script = task_info.Script
	}
	// XXX: To replace with PID handling and background process.
	// XXX: Exp. in docker container
	// XXX: Start with docker, monitor it.
	// XXX: optional args, with --priviledged and -v socket
	docker_client := d.DockerClient

	if len(task_info.Image) > 0 {

		if len(task_info.CacheImage) > 0 {

			if img, err := d.FindImage(sharedName); err == nil {
				fetcher.AppendTaskOutput("Cached image found: " + img + " " + sharedName)
				if len(task_info.CacheClean) > 0 {
					fetcher.AppendTaskOutput("Not using previously cached image - deleting image: " + sharedName)
					d.RemoveImage(sharedName)
				} else {
					image = img
				}
			} else {
				fetcher.AppendTaskOutput("No cached image found for '" + sharedName + "'")
			}
		}

		fetcher.AppendTaskOutput("Pulling image: " + task_info.Image)
		if err = docker_client.PullImage(docker.PullImageOptions{Repository: task_info.Image}, docker.AuthConfiguration{}); err != nil {
			panic(err)
		}
		fetcher.AppendTaskOutput("Pulling image: DONE!")
	}
	//var args []string
	var git_root_path = d.Context.RootTaskDir
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

	for _, b := range task_info.Binds {
		ContainerBinds = append(ContainerBinds, b)
	}

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

	if err := d.DownloadArtefacts(artefactdir, storagedir); err != nil {
		return 1, err
	}

	//ContainerVolumes = append(ContainerVolumes, git_repo_dir+":/build")

	ContainerBinds = append(ContainerBinds, git_repo_dir+":"+git_root_path)

	createContHostConfig := docker.HostConfig{
		Privileged: setting.Configuration.DockerPriviledged,
		Binds:      ContainerBinds,
		CapAdd:     setting.Configuration.DockerCaps,
		CapDrop:    setting.Configuration.DockerCapsDrop,
		//	LogConfig:  docker.LogConfig{Type: "json-file"}
	}
	var containerconfig = &docker.Config{Image: image, WorkingDir: git_build_root_path}

	if len(execute_script) > 0 {
		containerconfig.Cmd = []string{"-c", "pwd;ls -liah;" + execute_script}
		containerconfig.Entrypoint = []string{"/bin/sh"}
	}

	if len(task_info.Environment) > 0 {
		containerconfig.Env = task_info.Environment
		//	fetcher.AppendTaskOutput("Env: ")
		//	for _, e := range task_info.Environment {
		//		fetcher.AppendTaskOutput("- " + e)
		//	}
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
	defer d.CleanUpContainer(container.ID)
	if setting.Configuration.DockerKeepImg == false {
		defer d.RemoveImage(task_info.Image)
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

			err = d.UploadArtefacts(to_upload)

			fetcher.AppendTaskOutput("Container execution terminated")

			if len(task_info.CacheImage) > 0 {
				fetcher.AppendTaskOutput("Saving container")
				d.CommitImage(container.ID, sharedName, "latest")
			}

			if len(task_info.Prune) > 0 {
				fetcher.AppendTaskOutput("Pruning unused docker resources")
				d.Prune()
			}
			if err != nil {
				return 1, err
			}
			return c_data.State.ExitCode, nil
		}
	}

}

func (d *DockerExecutor) CommitImage(containerID, repo, tag string) (string, error) {
	image, err := d.DockerClient.CommitContainer(docker.CommitContainerOptions{Container: containerID, Repository: repo, Tag: tag})
	if err != nil {
		return "", err
	}
	return image.ID, nil
}

func (d *DockerExecutor) FindImage(image string) (string, error) {
	images, err := d.DockerClient.ListImages(docker.ListImagesOptions{Filter: image})
	if err != nil {
		return "", err
	}
	if len(images) > 0 {
		return images[0].ID, nil
	}
	return "", errors.New("Image not found")
}

func (d *DockerExecutor) RemoveImage(image string) error {
	return d.DockerClient.RemoveImage(image)
}

func (d *DockerExecutor) NewImageFrom(image, newimage, tag string) error {
	images, err := d.DockerClient.ListImages(docker.ListImagesOptions{Filter: image})
	var id string
	if len(images) > 0 {
		id = images[0].ID
	}

	err = d.DockerClient.TagImage(id, docker.TagImageOptions{Repo: newimage, Tag: tag, Force: true})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerExecutor) CleanUpContainer(ID string) error {
	return d.DockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ID,
		Force: true,
	})
}
