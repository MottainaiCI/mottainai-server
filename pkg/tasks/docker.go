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
	"path"
	"strings"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerExecutor struct {
	*TaskExecutor
	DockerClient *docker.Client
}

func NewDockerExecutor(config *setting.Config) *DockerExecutor {
	return &DockerExecutor{
		TaskExecutor: &TaskExecutor{
			Context: NewExecutorContext(),
			Config:  config,
		}}
}

func (e *DockerExecutor) Prune() {
	e.DockerClient.PruneContainers(docker.PruneContainersOptions{})
	e.DockerClient.PruneImages(docker.PruneImagesOptions{})
	e.DockerClient.PruneVolumes(docker.PruneVolumesOptions{})
	e.DockerClient.PruneNetworks(docker.PruneNetworksOptions{})
}

func (d *DockerExecutor) Setup(docID string) error {
	d.TaskExecutor.Setup(docID)
	docker_client, err := docker.NewClient(d.Config.GetAgent().DockerEndpoint)
	if err != nil {
		return (errors.New("Endpoint:" + d.Config.GetAgent().DockerEndpoint + " Error: " + err.Error()))
	}
	d.DockerClient = docker_client
	return nil
}

func purgeImageName(image string) string {
	return strings.Replace(image, "/", "-", -1)
}

func (d *DockerExecutor) Play(docID string) (int, error) {
	fetcher := d.MottainaiClient
	th := DefaultTaskHandler(d.Config)
	task_info := th.FetchTask(fetcher)
	image := task_info.Image

	sharedName, err := d.TaskExecutor.CreateSharedImageName(&task_info)
	if err != nil {
		return 1, err
	}

	artdir := d.Context.ArtefactDir
	storagetmp := d.Context.StorageDir
	git_repo_dir := d.Context.SourceDir

	var execute_script = "mottainai-run"

	if len(task_info.Script) > 0 {
		execute_script = strings.Join(task_info.Script, " && ")
	}
	// XXX: To replace with PID handling and background process.
	// XXX: Exp. in docker container
	// XXX: Start with docker, monitor it.
	// XXX: optional args, with --priviledged and -v socket
	docker_client := d.DockerClient

	if len(task_info.Image) > 0 {

		if len(task_info.CacheImage) > 0 {

			if img, err := d.FindImage(sharedName); err == nil {
				d.Report("Cached image found: " + img + " " + sharedName)
				if len(task_info.CacheClean) > 0 {
					d.Report("Not using previously cached image - deleting image: " + sharedName)
					d.RemoveImage(sharedName)
				} else {
					image = img
				}
			} else {
				d.Report("No cached image found locally for '" + sharedName + "'")
			}

			if len(task_info.CacheClean) == 0 {
				// Retrieve cached image into the hub
				if t, oktype := d.Config.GetAgent().CacheRegistryCredentials["type"]; oktype && t == "docker" {

					toPull := purgeImageName(sharedName)
					if e, ok := d.Config.GetAgent().CacheRegistryCredentials["entity"]; ok {
						toPull = e + "/" + toPull
					}
					d.Report("Try to pull cache (" + toPull + ") image from defined registry or from dockerhub")

					if baseUrl, okb := d.Config.GetAgent().CacheRegistryCredentials["baseurl"]; okb {
						toPull = baseUrl + toPull
					}

					if e := d.PullImage(toPull); e == nil {
						image = toPull
						d.Report("Using pulled image:  " + image)
					} else {
						d.Report("No image could be fetched by cache registry")
					}
				}
			}

		}

		d.Report("Pulling image: " + task_info.Image)
		if err := d.PullImage(task_info.Image); err != nil {
			return 1, err
		}
		d.Report("Pulling image: DONE!")
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

	if d.Config.GetAgent().DockerInDocker {
		ContainerBinds = append(ContainerBinds, d.Config.GetAgent().DockerEndpointDiD+":/var/run/docker.sock")
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
	if len(git_repo_dir) > 0 {
		ContainerBinds = append(ContainerBinds, git_repo_dir+":"+git_root_path)
	}

	createContHostConfig := docker.HostConfig{
		Privileged: d.Config.GetAgent().DockerPriviledged,
		Binds:      ContainerBinds,
		CapAdd:     d.Config.GetAgent().DockerCaps,
		CapDrop:    d.Config.GetAgent().DockerCapsDrop,
		//	LogConfig:  docker.LogConfig{Type: "json-file"}
	}
	var containerconfig = &docker.Config{Image: image, WorkingDir: git_build_root_path}
	d.Report("Execute: " + execute_script)
	if len(execute_script) > 0 {
		containerconfig.Cmd = []string{"-c", "pwd;ls -liah;" + execute_script}
		containerconfig.Entrypoint = []string{"/bin/bash"}
	}

	if len(task_info.Entrypoint) > 0 {
		containerconfig.Entrypoint = task_info.Entrypoint
		containerconfig.Cmd = task_info.Script
		d.Report("Entrypoint: " + strings.Join(containerconfig.Entrypoint, ","))
	}

	if len(task_info.Environment) > 0 {
		containerconfig.Env = task_info.Environment
		//	d.Report("Env: ")
		//	for _, e := range task_info.Environment {
		//		d.Report("- " + e)
		//	}
	}

	d.Report("Binds: ")
	for _, v := range ContainerBinds {
		d.Report("- " + v)
	}

	d.Report("Container working dir: " + git_build_root_path)
	d.Report("Image: " + containerconfig.Image)

	container, err := docker_client.CreateContainer(docker.CreateContainerOptions{
		Config:     containerconfig,
		HostConfig: &createContHostConfig,
	})

	if err != nil {
		panic(err)
	}

	utils.ContainerOutputAttach(func(s string) {
		d.Report(s)
	}, docker_client, container)
	defer d.CleanUpContainer(container.ID)
	if d.Config.GetAgent().DockerKeepImg == false {
		defer d.RemoveImage(task_info.Image)
	}

	d.Report("Created container ID: " + container.ID)

	err = docker_client.StartContainer(container.ID, &createContHostConfig)
	if err != nil {
		panic(err)
	}
	d.Report("Started Container " + container.ID)

	starttime := time.Now()

	for {
		time.Sleep(1 * time.Second)
		now := time.Now()
		task_info = th.FetchTask(fetcher)
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			return d.HandleTaskStop(timedout)
		}
		c_data, err := docker_client.InspectContainer(container.ID) // update our container information
		if err != nil {
			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			d.Report(err.Error())
			return 0, nil
		}
		if c_data.State.Running == false {

			var err error

			to_upload := artdir
			if d.Config.GetAgent().DockerInDocker {
				to_upload = path.Join(git_root_path, task_info.Directory, artefact_path)
			}

			err = d.UploadArtefacts(to_upload)
			if err != nil {
				return 1, err
			}
			d.Report("Container execution terminated")

			if len(task_info.CacheImage) > 0 {
				d.Report("Saving container to " + sharedName)
				d.CommitImage(container.ID, sharedName, "latest")

				// Push image, if a cache_registry is configured in the node
				if err := d.PushImage(sharedName); err != nil {
					d.Report("Failed pushing image to cache registry: " + err.Error())
				} else {
					d.Report("Image pushed to cache registry successfully")
				}
			}

			if len(task_info.Prune) > 0 {
				d.Report("Pruning unused docker resources")
				d.Prune()
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
	return d.DockerClient.RemoveImageExtended(image, docker.RemoveImageOptions{
		Force:   true,
		NoPrune: false,
	})
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
		ID:            ID,
		Force:         true,
		RemoveVolumes: true,
	})
}

func (d *DockerExecutor) PullImage(image string) error {

	username, okname := d.Config.GetAgent().CacheRegistryCredentials["username"]
	password, okpassword := d.Config.GetAgent().CacheRegistryCredentials["password"]
	auth := docker.AuthConfiguration{}
	if okname && okpassword {
		auth.Password = password
		auth.Username = username
	}
	if serveraddress, ok := d.Config.GetAgent().CacheRegistryCredentials["serveraddress"]; ok {
		auth.ServerAddress = serveraddress
	}
	d.MottainaiClient.AppendTaskOutput("Pulling image: " + image)
	if err := d.DockerClient.PullImage(docker.PullImageOptions{Repository: image}, auth); err != nil {
		return err
	}

	return nil
}

func (d *DockerExecutor) PushImage(image string) error {

	// Push image, if a cache_registry is configured in the node
	if t, oktype := d.Config.GetAgent().CacheRegistryCredentials["type"]; oktype && t == "docker" {

		username, okname := d.Config.GetAgent().CacheRegistryCredentials["username"]
		password, okpassword := d.Config.GetAgent().CacheRegistryCredentials["password"]

		if okname && okpassword {
			baseurl, okbaseurl := d.Config.GetAgent().CacheRegistryCredentials["baseurl"]
			entity, okentity := d.Config.GetAgent().CacheRegistryCredentials["entity"]
			imageopts := docker.PushImageOptions{}
			auth := docker.AuthConfiguration{}
			auth.Password = password
			auth.Username = username
			serveraddress, okserveraddress := d.Config.GetAgent().CacheRegistryCredentials["serveraddress"]
			if okserveraddress {
				auth.ServerAddress = serveraddress
			}
			if okentity {
				imageopts.Name = entity + "/" + purgeImageName(image)
				if err := d.NewImageFrom(image, imageopts.Name, "latest"); err != nil {
					return err
				}
				d.MottainaiClient.AppendTaskOutput("Tagged image: " + image + " ----> " + imageopts.Name)
			} else {
				imageopts.Name = purgeImageName(image)
			}
			d.MottainaiClient.AppendTaskOutput("Pushing image: " + imageopts.Name)

			if okbaseurl {
				imageopts.Registry = baseurl
			}
			return d.DockerClient.PushImage(imageopts, auth)
		}
	} else {
		return errors.New("No cache registry set - only local cache is available")
	}
	return nil
}

func (d *DockerExecutor) FindImageInHub(image string) (bool, error) {
	res, err := d.DockerClient.SearchImages(image)
	if err != nil {
		return false, err
	}
	for _, r := range res {
		if r.Name == image {
			return true, nil
		}
	}

	return false, nil
}
