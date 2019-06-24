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
	"strings"
	"time"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

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

func (d *DockerExecutor) HandleCacheImagePush(req StateRequest, task_info tasks.Task) {
	if len(task_info.CacheImage) > 0 {
		d.Report("Saving container to " + req.CacheImage)
		d.CommitImage(req.ContainerID, req.CacheImage, "latest")
		// Push image, if a cache_registry is configured in the node
		if err := d.PushImage(req.CacheImage); err != nil {
			d.Report("Failed pushing image to cache registry: " + err.Error())
		} else {
			d.Report("Image pushed to cache registry successfully")
		}
	}
}

func (d *DockerExecutor) ImageToCacheName(image string) string {

	if t, oktype := d.Config.GetAgent().CacheRegistryCredentials["type"]; oktype && t == "docker" {

		toPull := purgeImageName(image)
		if e, ok := d.Config.GetAgent().CacheRegistryCredentials["entity"]; ok {
			toPull = e + "/" + toPull
		}

		if baseUrl, okb := d.Config.GetAgent().CacheRegistryCredentials["baseurl"]; okb {
			toPull = baseUrl + toPull
		}

		return toPull
	}
	return image
}

func (d *DockerExecutor) ResolveCachedImage(task_info tasks.Task) (string, string, error) {
	image := task_info.Image
	// That's the image we will update in case caching is enabled
	sharedName, err := d.TaskExecutor.CreateSharedImageName(&task_info)
	if err != nil {
		return "", sharedName, err
	}

	if len(task_info.Image) > 0 {
		if len(task_info.CacheImage) > 0 {
			if img, err := d.FindImage(sharedName); err == nil {
				d.Report(">> Cached image found: " + img + " " + sharedName)
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
				toPull := d.ImageToCacheName(sharedName)

				d.Report("Try to pull cache (" + toPull + ") image from defined registry or from dockerhub")

				if e := d.PullImage(toPull); e == nil {
					image = toPull
					d.Report(">> Using pulled image:  " + image)
				} else {
					d.Report(">> No image could be fetched by cache registry")
				}
			}
		}

		d.Report(">> Pulling image: " + image)
		if err := d.PullImage(image); err != nil {
			return "", sharedName, err
		}
		d.Report(">> Pulling image: DONE!")
	}
	return image, sharedName, nil
}

func (d *DockerExecutor) AttachContainerReport(container *docker.Container) {
	utils.ContainerOutputAttach(func(s string) {
		d.Report(s)
	}, d.DockerClient, container)
}

func (d *DockerExecutor) Play(docID string) (int, error) {
	task_info, err := tasks.FetchTask(d.MottainaiClient)
	if err != nil {
		return 1, err
	}

	instruction := NewInstructionFromTask(task_info)

	d.Context.ResolveMounts(instruction)

	// That is the image we are using for the build
	image, cachedimage, err := d.ResolveCachedImage(task_info)
	if err != nil {
		return 1, err
	}

	mapping := d.Context.ResolveArtefactsMounts(ArtefactMapping{
		ArtefactPath: task_info.ArtefactPath,
		StoragePath:  task_info.StoragePath,
	}, instruction, d.Config.GetAgent().DockerInDocker)

	if d.Config.GetAgent().DockerInDocker {
		instruction.AddMount(d.Config.GetAgent().DockerEndpointDiD + ":/var/run/docker.sock")
	}

	instruction.Report(d)
	d.Context.Report(d)

	if err := d.DownloadArtefacts(mapping.ArtefactPath, mapping.StoragePath); err != nil {
		return 1, err
	}
	d.Report(">> Creating container..")

	container, err := d.DockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      image,
			WorkingDir: d.Context.HostPath(task_info.Directory),
			Cmd:        instruction.CommandList(),
			Entrypoint: instruction.EntrypointList(),
			Env:        instruction.EnvironmentList(),
		},
		HostConfig: &docker.HostConfig{
			Privileged: d.Config.GetAgent().DockerPriviledged,
			Binds:      instruction.MountsList(),
			CapAdd:     d.Config.GetAgent().DockerCaps,
			CapDrop:    d.Config.GetAgent().DockerCapsDrop,
		},
	})
	if err != nil {
		d.Report("Creating container error: " + err.Error())
		return 1, err
	}
	d.AttachContainerReport(container)
	d.Report("Created container ID: " + container.ID)
	request := StateRequest{
		ContainerID:   container.ID,
		ImagesToClean: []string{image},
		CacheImage:    cachedimage,
		Prune:         len(task_info.Prune) > 0,
	}

	defer d.CleanUpContainer(request)

	// FIXME: Replace with goroutine?
	err = d.DockerClient.StartContainer(container.ID, container.HostConfig)
	if err != nil {
		d.Report("Starting container error: " + err.Error())
		return 1, err
	}
	d.Report(">> Started Container " + container.ID)

	// We always update the cache image
	return d.Handle(request, mapping)
}

func (d *DockerExecutor) Handle(req StateRequest, mapping ArtefactMapping) (int, error) {
	starttime := time.Now()

	for {
		time.Sleep(1 * time.Second)
		now := time.Now()
		task_info, err := tasks.FetchTask(d.MottainaiClient)
		if err != nil {
			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			d.Report(err.Error())
			return 0, nil
		}
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			return d.HandleTaskStop(timedout)
		}
		c_data, err := d.DockerClient.InspectContainer(req.ContainerID) // update our container information
		if err != nil {
			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			d.Report(err.Error())
			return 0, nil
		}
		if c_data.State.Running == false {
			d.Report("Container execution terminated")

			d.Report("Upload of artifacts starts")
			err := d.UploadArtefacts(mapping.ArtefactPath)
			if err != nil {
				return 1, err
			}
			d.Report("Upload of artifacts terminated")

			d.HandleCacheImagePush(req, task_info)

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

func (d *DockerExecutor) CleanUpContainer(req StateRequest) error {
	d.Report("Cleanup container")

	err := d.DockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            req.ContainerID,
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		d.Report("Container cleanup error: ", err.Error())
	}

	if d.Config.GetAgent().DockerKeepImg == false {
		for _, i := range req.ImagesToClean {
			d.Report("Removing image " + i)
			d.RemoveImage(i)
		}
	}

	if req.Prune {
		d.Report("Pruning unused docker resources")
		d.Prune()
	}

	return err
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
			imageopts := docker.PushImageOptions{}
			auth := docker.AuthConfiguration{}
			auth.Password = password
			auth.Username = username
			serveraddress, okserveraddress := d.Config.GetAgent().CacheRegistryCredentials["serveraddress"]

			imageopts.Name = d.ImageToCacheName(image)

			if okserveraddress {
				auth.ServerAddress = serveraddress
			}

			if d.ImageToCacheName(image) != image {
				if err := d.NewImageFrom(image, imageopts.Name, "latest"); err != nil {
					return err
				}
				d.MottainaiClient.AppendTaskOutput("Tagged image: " + image + " ----> " + imageopts.Name)
			}

			if okbaseurl {
				imageopts.Registry = baseurl
			}

			d.MottainaiClient.AppendTaskOutput("Pushing image: " + imageopts.Name)

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
