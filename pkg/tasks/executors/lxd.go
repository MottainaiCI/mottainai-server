// +build lxd

/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>
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
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"container/list"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	lxd "github.com/lxc/lxd/client"
	lxd_config "github.com/lxc/lxd/lxc/config"
	lxd_utils "github.com/lxc/lxd/lxc/utils"
	lxd_shared "github.com/lxc/lxd/shared"
	lxd_api "github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/ioprogress"
)

func NewLxdExecutor(config *setting.Config) *LxdExecutor {
	return &LxdExecutor{
		TaskExecutor: &TaskExecutor{
			Context: NewExecutorContext(),
			Config:  config,
		},
	}
}

type LxdExecutor struct {
	*TaskExecutor
	LxdClient lxd.ContainerServer
	LxdConfig *lxd_config.Config
	// Required for handle cancellable task
	CurrentLocalOperation lxd.Operation
	RemoteOperation       lxd.RemoteOperation
}

func (e *LxdExecutor) Prune() {
	// TODO
}

func (l *LxdExecutor) Setup(docID string) error {
	l.TaskExecutor.Setup(docID)

	var err error
	var client lxd.ContainerServer
	var configPath string = path.Join(l.Config.GetAgent().BuildPath, "/lxc/config.yml")

	if len(l.Config.GetAgent().LxdConfigDir) > 0 {
		// TODO: handle path
		configPath = path.Join(l.Config.GetAgent().LxdConfigDir, "/config.yml")
	}

	l.LxdConfig, err = lxd_config.LoadConfig(configPath)
	if err != nil {
		return (errors.New("Error on load LXD config: " + err.Error()))
	}

	if len(l.Config.GetAgent().LxdEndpoint) > 0 {
		client, err = lxd.ConnectLXDUnix(l.Config.GetAgent().LxdEndpoint, nil)
		if err != nil {
			return (errors.New("Endpoint:" + l.Config.GetAgent().LxdEndpoint + " Error: " + err.Error()))
		}
		// Force use of local
		l.LxdConfig.DefaultRemote = "local"
	} else {
		if len(l.LxdConfig.DefaultRemote) > 0 {
			// POST: If is present default I use default as main ContainerServer
			client, err = l.LxdConfig.GetContainerServer(l.LxdConfig.DefaultRemote)
		} else {
			if _, has_local := l.LxdConfig.Remotes["local"]; has_local {
				client, err = l.LxdConfig.GetContainerServer("local")
				// POST: I use local if is present
			} else {
				// POST: I use default socket connection
				client, err = lxd.ConnectLXDUnix("", nil)
			}
			if err != nil {
				return (errors.New("Error on create LXD Connector: " + err.Error()))
			}

			l.LxdConfig.DefaultRemote = "local"
		}
	}

	l.LxdClient = client

	return nil
}

func (l *LxdExecutor) ResolveCachedImage(task_info tasks.Task) (string, bool, string, error) {
	var cachedImage bool = false
	var imageFingerprint string

	sharedName, err := l.TaskExecutor.CreateSharedImageName(&task_info)
	if err != nil {
		return "", cachedImage, sharedName, err
	}

	if len(task_info.Image) > 0 {

		// NOTE: If cache image is enable and cache clean is disable then i
		//       search for image with sharedName between all remotes and I use
		//       fingerprint instead of task_info.Image.
		//       If cache image is enable and cache clean is enable I ignore
		//       caching because at the end of the process I delete aliases from
		//       existing image (if existing in local remote).
		if len(task_info.CacheImage) > 0 {
			cachedImage = true

			if len(task_info.CacheClean) == 0 {

				if imageFingerprint, _, _, _ = l.FindImage(sharedName); imageFingerprint != "" {
					l.Report("Cached image found: " + imageFingerprint + " " + sharedName)
					if imageFingerprint, err = l.PullImage(imageFingerprint); err != nil {
						return "", cachedImage, sharedName, err
					}
					l.Report(fmt.Sprintf("Pulling image %s: DONE!", imageFingerprint))
				}
			}
		}

		if imageFingerprint == "" {
			// POST: Cache image is disable or not caching image found.
			if imageFingerprint, err = l.PullImage(task_info.Image); err != nil {
				return "", cachedImage, sharedName, err
			}
			l.Report(fmt.Sprintf("Pulling image %s: DONE!", imageFingerprint))
		}
	}

	return imageFingerprint, cachedImage, sharedName, nil
}

func (l *LxdExecutor) InitContainer(containerName string, mapping ArtefactMapping) (string, error) {
	// Push workdir to container
	var err error
	var localWorkDir, targetHomeDir string

	targetHomeDir = strings.TrimRight(l.Context.ContainerPath(""), "/") + "/"

	if len(l.Context.SourceDir) > 0 {
		// NOTE: Last / it's needed to avoid to drop last directory on push directory
		localWorkDir = strings.TrimRight(l.Context.SourceDir, "/") + "/"
	} else {
		// NOTE: I use BuildDir to avoid execution of low level mkdir command
		//       on container. We can replace this with a mkdir to target
		localWorkDir = strings.TrimRight(l.Context.BuildDir, "/") + "/"
	}

	// Create workdir on container
	// Inside container I use the same path configured on agent with task id
	err = l.RecursivePushFile(containerName, localWorkDir, l.Context.RootTaskDir)
	if err != nil {
		return "", err
	}
	// Create artefactdir on container
	err = l.RecursivePushFile(containerName,
		strings.TrimRight(l.Context.ArtefactDir, "/")+"/", mapping.ArtefactPath)
	if err != nil {
		return "", err
	}
	// Create storagedir on container
	err = l.RecursivePushFile(containerName,
		strings.TrimRight(l.Context.StorageDir, "/")+"/", mapping.StoragePath)
	if err != nil {
		return "", err
	}

	return targetHomeDir, nil
}

func (l *LxdExecutor) Play(docId string) (int, error) {
	var imageFingerprint, containerName, sharedName string
	var cachedImage bool = false
	var err error

	task_info, err := tasks.FetchTask(l.MottainaiClient)
	if err != nil {
		return 1, err
	}

	imageFingerprint, cachedImage, sharedName, err = l.ResolveCachedImage(task_info)
	if err != nil {
		return 1, err
	}

	mapping := l.Context.ResolveArtefacts(ArtefactMapping{
		ArtefactPath: task_info.ArtefactPath,
		StoragePath:  task_info.StoragePath,
	})

	l.Context.Report(l)

	if err := l.DownloadArtefacts(l.Context.ArtefactDir, l.Context.StorageDir); err != nil {
		return 1, err
	}

	l.Report("Completed phase of artefacts download!")

	containerName = l.GetContainerName(&task_info)

	l.Report(">> Creating container " + containerName + "...")
	err = l.LaunchContainer(containerName, imageFingerprint, cachedImage)
	if err != nil {
		l.Report("Creating container error: " + err.Error())
		return 1, err
	}
	defer l.CleanUpContainer(containerName, &task_info)

	// Push workdir to container
	var targetHomeDir string

	targetHomeDir, err = l.InitContainer(containerName, mapping)
	if err != nil {
		l.Report("Error on initialize container: " + err.Error())
		return 1, err
	}

	exec := StateExecution{
		Request: &StateRequest{
			ContainerID: containerName,
			CacheImage:  sharedName,
			Prune:       false,
		},
		Status: "prepare",
		Result: -1,
	}

	// Run execution in background
	go l.ExecCommand(&exec, targetHomeDir, &task_info)

	return l.Handle(&exec, mapping)
}

func (l *LxdExecutor) Handle(exec *StateExecution, mapping ArtefactMapping) (int, error) {
	var err error
	var task_info tasks.Task

	starttime := time.Now()

	// Clean local ArtefactDir to retrieve result files
	if err = os.RemoveAll(l.Context.ArtefactDir); err != nil {
		l.Report("Warn: (cleanup failed on " + l.Context.ArtefactDir + ") : " + err.Error())
	}

	for {
		time.Sleep(1 * time.Second)
		now := time.Now()
		task_info, err = tasks.FetchTask(l.MottainaiClient)
		if err != nil {
			l.Report(err.Error())
			return 0, nil
		}
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))

		if task_info.IsStopped() || timedout {
			if exec.Status == "prepare" {
				// Wait until execution is on state running
				continue
			}

			// Stop container. I can't stop running lxd exec command
			err = l.DoAction2Container(exec.Request.ContainerID, "stop")
			if err != nil {
				l.Report("Error on stop container: " + err.Error())
				return 1, err
			}
			return l.HandleTaskStop(timedout)
		}

		if exec.Status != "prepare" && exec.Status != "running" {
			break
		}

	} // end for

	l.Report("Container execution terminated")

	// Pull ArtefactDir from container
	err = l.RecursivePullFile(exec.Request.ContainerID, mapping.ArtefactPath,
		l.Context.ArtefactDir, true)
	if err != nil {
		return 1, err
	}

	l.Report("Upload of artifacts starts")
	err = l.UploadArtefacts(l.Context.ArtefactDir)
	if err != nil {
		return 1, err
	}
	l.Report("Upload of artifacts terminated")

	if exec.Status == "error" {
		return exec.Result, exec.Error
	}

	err = l.HandleCacheImagePush(exec, mapping, &task_info)
	if err != nil {
		return 1, err
	}

	return exec.Result, nil
}

// Push image to a specific remote notes (use LXD protocol)
func (l *LxdExecutor) PushImage(fingerprint string, alias string) error {

	var err error
	var image_server lxd.ContainerServer

	remote, okremote := l.Config.GetAgent().LxdCacheRegistry["remote"]
	if !okremote {
		return fmt.Errorf("No remote param found under lxd_cache_registry.")
	}

	if remote == l.LxdConfig.DefaultRemote {
		l.Report("Remote server is equal to local. Nothing to push.")
		return nil
	}

	image_server, err = l.LxdConfig.GetContainerServer(remote)
	if err != nil {
		return fmt.Errorf("Error on retrieve remote %s: %s", remote, err)
	}

	// Check if an image with same alias is already present.
	aliasEntry, _, _ := image_server.GetImageAlias(alias)
	if aliasEntry != nil {
		image, _, _ := image_server.GetImage(aliasEntry.Target)
		if image != nil {
			err = image_server.DeleteImageAlias(alias)
			if err != nil {
				l.Report(fmt.Sprintf(
					"WARNING: Error on delete alias to remote image %s. I try to proceed.",
					aliasEntry.Target))
			}
		}
	}

	err = l.CopyImage(fingerprint, l.LxdClient, image_server)
	if err != nil {
		return fmt.Errorf("Error on copy image %s: %s", fingerprint, err)
	}

	return nil
}

// Create image on local LXD instance from an active container.
func (l *LxdExecutor) CommitImage(containerName, aliasName string, task *tasks.Task) (string, error) {

	var description string
	var err error

	// TODO: Check if enable Expires on image created.

	// Initialize properties for images
	properties := map[string]string{}

	// Check if there is already a local image with same alias. If yes I drop alias.
	aliasEntry, _, _ := l.LxdClient.GetImageAlias(aliasName)
	if aliasEntry != nil {
		l.Report(fmt.Sprintf(
			"Found old image %s with alias %s. I drop alias from it.",
			aliasEntry.Target, aliasName))

		err = l.LxdClient.DeleteImageAlias(aliasName)
		if err != nil {
			return "", err
		}
	}

	if task.Source != "" {
		description = fmt.Sprintf("Mottainai generated Image from %s for %s",
			task.Image, task.Source)
		properties["source"] = task.Image
	} else {
		description = fmt.Sprintf("Mottainai generated Image from %s", task.Image)
	}

	properties["description"] = description
	if task.Directory != "" {
		properties["directory"] = task.Directory
	}

	// Reformat aliases
	alias := lxd_api.ImageAlias{}
	alias.Name = aliasName
	aliases := []lxd_api.ImageAlias{alias}

	compression, okcompression := l.Config.GetAgent().LxdCacheRegistry["compression_algorithm"]
	if !okcompression {
		compression = "none"
	}

	// Create the image
	req := lxd_api.ImagesPost{
		Source: &lxd_api.ImagesPostSource{
			Type: "container",
			Name: containerName,
		},
		// CompressionAlgorithm contains name of the binary called by LXD for compression.
		// For any customization create custom script that wrap compression tools.
		CompressionAlgorithm: compression,
	}
	req.Properties = properties
	req.Public = true

	// TODO: Take time and calculate how much time is required for create image
	l.Report(fmt.Sprintf("Starting creation of Image with alias %s...", aliasName))

	l.CurrentLocalOperation, err = l.LxdClient.CreateImage(req, nil)
	if err != nil {
		return "", err
	}

	err = l.waitOperation(nil, nil)
	if err != nil {
		return "", err
	}

	opAPI := l.CurrentLocalOperation.Get()

	// Grab the fingerprint
	fingerprint := opAPI.Metadata["fingerprint"].(string)

	// Get the source image
	_, _, err = l.LxdClient.GetImage(fingerprint)
	if err != nil {
		return "", err
	}

	l.Report(fmt.Sprintf(
		"For container %s created image %s. Adding alias %s to image.",
		containerName, fingerprint, aliasName))

	for _, alias := range aliases {
		aliasPost := lxd_api.ImageAliasesPost{}
		aliasPost.Name = alias.Name
		aliasPost.Target = fingerprint
		err := l.LxdClient.CreateImageAlias(aliasPost)
		if err != nil {
			return "", fmt.Errorf("Failed to create alias %s", alias.Name)
		}
	}

	return fingerprint, nil
}

func (l *LxdExecutor) CleanUpContainer(containerName string, task *tasks.Task) error {
	var err error

	err = l.DoAction2Container(containerName, "stop")
	if err != nil {
		l.Report("Error on stop container: " + err.Error())
		return err
	}

	if len(task.CacheImage) > 0 && l.Config.GetAgent().LxdEphemeralContainers {
		// Delete container
		l.CurrentLocalOperation, err = l.LxdClient.DeleteContainer(containerName)
		if err != nil {
			l.Report("Error on delete container: " + err.Error())
			return err
		}
		_ = l.waitOperation(nil, nil)
	}

	return nil
}

func (l *LxdExecutor) HandleCacheImagePush(exec *StateExecution, mapping ArtefactMapping, task_info *tasks.Task) error {
	var err error
	var imageFingerprint string

	containerName := exec.Request.ContainerID
	sharedName := exec.Request.CacheImage

	if len(task_info.CacheImage) > 0 {

		l.Report("Try to clean artefacts and storage directories from container before create cached image...")

		// Delete old directories of storage and artefacts
		err = l.DeleteContainerDirRecursive(containerName, mapping.ArtefactPath)
		if err != nil {
			l.Report("WARNING: Error on clean artefacts dir on container: " + err.Error())
			// Ignore error. I'm not sure that is the right thing.
		}

		err = l.DeleteContainerDirRecursive(containerName, mapping.StoragePath)
		if err != nil {
			l.Report("WARNING: Error on clean storage dir on container: " + err.Error())
			// Ignore error. I'm not sure that is the right thing.
		}

		// Stop container for create image.
		err = l.DoAction2Container(containerName, "stop")
		if err != nil {
			l.Report("Error on stop container: " + err.Error())
			return err
		}

		l.Report("Saving container with alias " + sharedName)
		imageFingerprint, err = l.CommitImage(containerName, sharedName, task_info)
		if err != nil {
			return err
		}

		crType, okType := l.Config.GetAgent().LxdCacheRegistry["type"]
		if !okType || crType == "" {
			// If cache_registry type is not present i use p2p mode.
			crType = "p2p"
		} else if crType != "p2p" && crType != "server" {
			l.Report("WARNING: Found invalid cache_registry type. I force p2p.")
			crType = "p2p"
		}

		if crType == "server" {
			// Push image to remote
			if err = l.PushImage(imageFingerprint, sharedName); err != nil {
				l.Report("Failed pushing image to cache registry: " + err.Error())
			} else {
				l.Report("Image pushed to cache registry successfully")
			}
		}

	} else {

		l.Report("Container execution terminated")

	}

	return nil
}

func (l *LxdExecutor) LaunchContainer(name, fingerprint string, cachedImage bool) error {

	var err error
	var image *lxd_api.Image
	var profiles []string = []string{}
	var opInfo *lxd_api.Operation

	if len(l.Config.GetAgent().LxdProfiles) > 0 {
		for _, profile := range l.Config.GetAgent().LxdProfiles {
			profiles = append(profiles, profile)
		}
	} else {
		profiles = append(profiles, "default")
	}

	// Note: Avoid to create devece map for root /. We consider to handle this
	//       as profile. Same for different storage.
	devicesMap := map[string]map[string]string{}
	configMap := map[string]string{}

	// Setup container creation request
	req := lxd_api.ContainersPost{
		Name: name,
	}
	req.Config = configMap
	req.Devices = devicesMap
	req.Profiles = profiles

	if cachedImage {
		// If cache image is enable i need container when call stop action.
		req.Ephemeral = false
	} else {
		req.Ephemeral = l.Config.GetAgent().LxdEphemeralContainers
	}

	// Retrieve image info
	image, _, err = l.LxdClient.GetImage(fingerprint)
	if err != nil {
		return err
	}

	//req.Source.Alias = alias

	// Create the container
	l.RemoteOperation, err = l.LxdClient.CreateContainerFromImage(l.LxdClient, *image, req)
	if err != nil {
		return err
	}
	// Watch the background operation
	progress := lxd_utils.ProgressRenderer{
		Format: "Retrieving image: %s",
		Quiet:  false,
	}

	_, err = l.RemoteOperation.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		return err
	}

	err = l.waitOperation(nil, &progress)
	if err != nil {
		progress.Done("")
		return err
	}
	progress.Done("")

	// Extract the container name
	opInfo, err = l.RemoteOperation.GetTarget()
	if err != nil {
		return err
	}

	containers, ok := opInfo.Resources["containers"]
	if !ok || len(containers) == 0 {
		return fmt.Errorf("didn't get any affected image, container or snapshot from server")
	}

	l.RemoteOperation = nil

	// Start container
	return l.DoAction2Container(name, "start")
}

func (l *LxdExecutor) DoAction2Container(name, action string) error {
	var err error
	var container *lxd_api.Container
	var operation lxd.Operation

	container, _, err = l.LxdClient.GetContainer(name)
	if err != nil {
		if action == "stop" {
			l.Report(fmt.Sprintf("Container %s not found. Already stopped nothing to do.", name))
			return nil
		}
		return err
	}

	if action == "start" && container.Status == "Started" {
		l.Report(fmt.Sprintf("Container %s is already started!", name))
		return nil
	} else if action == "stop" && container.Status == "Stopped" {
		l.Report(fmt.Sprintf("Container %s is already stopped!", name))
		return nil
	}

	l.Report(fmt.Sprintf(
		"Trying to execute action %s to container %s: %v",
		action, name, container,
	))

	req := lxd_api.ContainerStatePut{
		Action:   action,
		Timeout:  120,
		Force:    false,
		Stateful: false,
	}

	operation, err = l.LxdClient.UpdateContainerState(name, req, "")
	if err != nil {
		l.Report("Error on update container state: " + err.Error())
		return err
	}

	progress := lxd_utils.ProgressRenderer{
		Quiet: false,
	}

	_, err = operation.AddHandler(progress.UpdateOp)
	if err != nil {
		l.Report("Error on add handler to progress bar: " + err.Error())
		progress.Done("")
		return err
	}

	err = l.waitOperation(operation, &progress)
	progress.Done("")
	if err != nil {
		l.Report(fmt.Sprintf("Error on stop container %s: %s", name, err))
		return err
	}

	if action == "start" {
		l.Report(fmt.Sprintf("Container %s is started!", name))
	} else {
		l.Report(fmt.Sprintf("Container %s is stopped!", name))
	}

	return nil
}

func (l *LxdExecutor) FindImage(image string) (string, lxd.ImageServer, string, error) {
	var err error
	var tmp_srv, srv lxd.ImageServer
	var img, tmp_img *lxd_api.Image
	var fingerprint string = ""
	var srv_name string = ""

	for remote, server := range l.LxdConfig.Remotes {
		tmp_srv, err = l.LxdConfig.GetImageServer(remote)
		if err != nil {
			err = nil
			l.Report(fmt.Sprintf(
				"Error on retrieve ImageServer for remote %s at addr %s",
				remote, server.Addr,
			))
			continue
		}
		tmp_img, err = l.GetImage(image, tmp_srv)
		if err != nil {
			// POST: No image found with input alias/fingerprint.
			//       I go ahead to next remote
			err = nil
			continue
		}

		if img != nil {
			// POST: A previous image is already found
			if tmp_img.CreatedAt.After(img.CreatedAt) {
				img = tmp_img
				srv = tmp_srv
				srv_name = remote
				fingerprint = img.Fingerprint
			}
		} else {
			// POST: first image matched
			img = tmp_img
			fingerprint = img.Fingerprint
			srv = tmp_srv
			srv_name = remote
		}
	}

	if fingerprint == "" {
		err = fmt.Errorf("No image found with alias or fingerprint %s", image)
	}

	return fingerprint, srv, srv_name, err
}

// Retrieve Image from alias or fingerprint to a specific remote.
func (l *LxdExecutor) GetImage(image string, remote lxd.ImageServer) (*lxd_api.Image, error) {
	var err error
	var img *lxd_api.Image
	var aliasEntry *lxd_api.ImageAliasesEntry

	img, _, err = remote.GetImage(image)
	if err != nil {
		// POST: no image found with input fingerprint
		//       Try to search an image as alias.

		// Check if exists an image with input alias
		aliasEntry, _, err = remote.GetImageAlias(image)
		if err != nil {
			img = nil
		} else {
			// POST: Find image with alias and so I try to retrieve api.Image
			//       object with all information.
			img, _, err = remote.GetImage(aliasEntry.Target)
		}
	}

	return img, err
}

// Delete alias from image of a specific ContainerServer if available
func (l *LxdExecutor) DeleteImageAliases4Alias(imageAlias string, server lxd.ContainerServer) error {
	var err error
	var img *lxd_api.Image

	img, _ = l.GetImage(imageAlias, server)
	if img != nil {
		err = l.DeleteImageAliases(img, server)
	}

	return err
}

// Delete all local alias defined on input Image to avoid conflict on pull.
func (l *LxdExecutor) DeleteImageAliases(image *lxd_api.Image, server lxd.ContainerServer) error {
	for _, alias := range image.Aliases {
		// Retrieve image with alias
		aliasEntry, _, _ := server.GetImageAlias(alias.Name)
		if aliasEntry != nil {
			// TODO: See how handle correctly this use case
			l.Report(fmt.Sprintf(
				"Found old image %s with alias %s. I drop alias from it.",
				aliasEntry.Target, alias.Name))

			err := server.DeleteImageAlias(alias.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *LxdExecutor) CopyImage(imageFingerprint string, remote lxd.ImageServer, to lxd.ContainerServer) error {
	var err error

	// Get the image information
	i, _, err := remote.GetImage(imageFingerprint)
	if err != nil {
		return err
	}

	copyArgs := &lxd.ImageCopyArgs{
		CopyAliases: true,
		Public:      true,
		AutoUpdate:  false,
	}

	// Ask LXD to copy the image from the remote server
	// CopyImage return an lxd.RemoteOperation does not implement lxd.Operation
	// (missing Cancel method) so DownloadImage is not s
	l.RemoteOperation, err = to.CopyImage(remote, *i, copyArgs)
	if err != nil {
		l.Report("Error on create copy image task " + err.Error())
		return err
	}

	// Watch the background operation
	progress := lxd_utils.ProgressRenderer{
		Format: "Retrieving image: %s",
		Quiet:  false,
	}

	_, err = l.RemoteOperation.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		l.RemoteOperation = nil
		return err
	}

	err = l.waitOperation(nil, &progress)
	progress.Done("")
	l.RemoteOperation = nil
	if err != nil {
		l.Report("Error on copy image " + err.Error())
		return err
	}

	l.Report(fmt.Sprintf("Image %s copy locally.", imageFingerprint))
	return nil
}

func (l *LxdExecutor) DownloadImage(imageFingerprint string, remote lxd.ImageServer) error {
	return l.CopyImage(imageFingerprint, remote, l.LxdClient)
}

func (l *LxdExecutor) PullImage(imageAlias string) (string, error) {
	var err error
	var imageFingerprint, remote_name string
	var remote lxd.ImageServer

	l.Report("Searching image: " + imageAlias)

	// Find image hashing id
	imageFingerprint, remote, remote_name, err = l.FindImage(imageAlias)
	if err != nil {
		return "", err
	}

	if imageFingerprint == imageAlias {
		l.Report("Use directly fingerprint " + imageAlias)
	} else {
		l.Report("For image " + imageAlias + " found fingerprint " + imageFingerprint)
	}

	// Check if image is already present locally else we receive an error.
	image, _, _ := l.LxdClient.GetImage(imageFingerprint)
	if image == nil {
		// NOTE: In concurrency could be happens that different image that
		//       share same aliases generate reset of aliases but
		//       if I work with fingerprint after FindImage I can ignore
		//       aliases.

		// Delete local image with same target aliases to avoid error on pull.
		err = l.DeleteImageAliases4Alias(imageAlias, l.LxdClient)

		// Try to pull image to lxd instance
		l.Report(fmt.Sprintf(
			"Try to download image %s from remote %s...",
			imageFingerprint, remote_name,
		))
		err = l.DownloadImage(imageFingerprint, remote)
	} else {
		l.Report("Image " + imageFingerprint + " already present.")
	}

	return imageFingerprint, err
}

// Based on code of lxc client tool https://github.com/lxc/lxd/blob/master/lxc/file.go
func (l *LxdExecutor) RecursiveMkdir(nameContainer string, dir string, mode *os.FileMode, uid int64, gid int64) error {

	/* special case, every container has a /, we don't need to do anything */
	if dir == "/" {
		return nil
	}

	// Remove trailing "/" e.g. /A/B/C/. Otherwise we will end up with an
	// empty array entry "" which will confuse the Mkdir() loop below.
	pclean := filepath.Clean(dir)
	parts := strings.Split(pclean, "/")
	i := len(parts)

	for ; i >= 1; i-- {
		cur := filepath.Join(parts[:i]...)
		_, resp, err := l.LxdClient.GetContainerFile(nameContainer, cur)
		if err != nil {
			continue
		}

		if resp.Type != "directory" {
			return fmt.Errorf("%s is not a directory", cur)
		}

		i++
		break
	}

	for ; i <= len(parts); i++ {
		cur := filepath.Join(parts[:i]...)
		if cur == "" {
			continue
		}

		cur = "/" + cur

		modeArg := -1
		if mode != nil {
			modeArg = int(mode.Perm())
		}
		args := lxd.ContainerFileArgs{
			UID:  uid,
			GID:  gid,
			Mode: modeArg,
			Type: "directory",
		}

		l.Report(fmt.Sprintf("Creating %s (%s)\n", cur, args.Type))

		err := l.LxdClient.CreateContainerFile(nameContainer, cur, args)
		if err != nil {
			return err
		}
	}

	return nil
}

// Based on code of lxc client tool https://github.com/lxc/lxd/blob/master/lxc/file.go
func (l *LxdExecutor) RecursivePushFile(nameContainer, source, target string) error {

	// Determine the target mode
	mode := os.FileMode(0755)
	// Create directory as root. TODO: see if we can use a specific user.
	var uid int64 = 0
	var gid int64 = 0
	err := l.RecursiveMkdir(nameContainer, target, &mode, uid, gid)
	if err != nil {
		return err
	}

	//source = filepath.Clean(source)
	//sourceDir, _ := filepath.Split(source)
	sourceDir := filepath.Clean(source)
	sourceLen := len(sourceDir)

	sendFile := func(p string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Failed to walk path for %s: %s", p, err)
		}

		// Detect unsupported files
		if !fInfo.Mode().IsRegular() && !fInfo.Mode().IsDir() && fInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			return fmt.Errorf("'%s' isn't a supported file type", p)
		}

		// Prepare for file transfer
		targetPath := path.Join(target, filepath.ToSlash(p[sourceLen:]))
		mode, uid, gid := lxd_shared.GetOwnerMode(fInfo)
		args := lxd.ContainerFileArgs{
			UID:  int64(uid),
			GID:  int64(gid),
			Mode: int(mode.Perm()),
		}

		var readCloser io.ReadCloser

		if fInfo.IsDir() {
			// Directory handling
			args.Type = "directory"
		} else if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			// Symlink handling
			symlinkTarget, err := os.Readlink(p)
			if err != nil {
				return err
			}

			args.Type = "symlink"
			args.Content = bytes.NewReader([]byte(symlinkTarget))
			readCloser = ioutil.NopCloser(args.Content)
		} else {
			// File handling
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()

			args.Type = "file"
			args.Content = f
			readCloser = f
		}

		progress := lxd_utils.ProgressRenderer{
			Format: fmt.Sprintf("Pushing %s to %s: %%s", p, targetPath),
			Quiet:  false,
		}

		if args.Type != "directory" {
			contentLength, err := args.Content.Seek(0, io.SeekEnd)
			if err != nil {
				return err
			}

			_, err = args.Content.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}

			args.Content = lxd_shared.NewReadSeeker(&ioprogress.ProgressReader{
				ReadCloser: readCloser,
				Tracker: &ioprogress.ProgressTracker{
					Length: contentLength,
					Handler: func(percent int64, speed int64) {

						l.Report(fmt.Sprintf("%d%% (%s/s)", percent,
							lxd_shared.GetByteSizeString(speed, 2)))

						progress.UpdateProgress(ioprogress.ProgressData{
							Text: fmt.Sprintf("%d%% (%s/s)", percent,
								lxd_shared.GetByteSizeString(speed, 2))})
					},
				},
			}, args.Content)
		}

		l.Report(fmt.Sprintf("Pushing %s to %s (%s)\n", p, targetPath, args.Type))
		err = l.LxdClient.CreateContainerFile(nameContainer, targetPath, args)
		if err != nil {
			if args.Type != "directory" {
				progress.Done("")
			}
			return err
		}
		if args.Type != "directory" {
			progress.Done("")
		}
		return nil
	}

	return filepath.Walk(source, sendFile)
}

// Based on code of lxc client tool https://github.com/lxc/lxd/blob/master/lxc/file.go
func (l *LxdExecutor) RecursivePullFile(nameContainer string, destPath string, localPath string, localAsTarget bool) error {

	buf, resp, err := l.LxdClient.GetContainerFile(nameContainer, destPath)
	if err != nil {
		return err
	}

	var target string
	// Default loging is to append tree to target directory
	if localAsTarget {
		target = localPath
	} else {
		target = filepath.Join(localPath, filepath.Base(destPath))
	}
	//target := localPath
	l.Report(fmt.Sprintf("Pulling %s from %s (%s)\n", target, destPath, resp.Type))

	if resp.Type == "directory" {
		err := os.MkdirAll(target, os.FileMode(resp.Mode))
		if err != nil {
			l.Report(fmt.Sprintf("directory %s is already present. Nothing to do.\n", target))
		}

		for _, ent := range resp.Entries {
			nextP := path.Join(destPath, ent)

			err = l.RecursivePullFile(nameContainer, nextP, target, false)
			if err != nil {
				return err
			}
		}
	} else if resp.Type == "file" {
		f, err := os.Create(target)
		if err != nil {
			return err
		}

		defer f.Close()

		err = os.Chmod(target, os.FileMode(resp.Mode))
		if err != nil {
			return err
		}

		progress := lxd_utils.ProgressRenderer{
			Format: fmt.Sprintf("Pulling %s from %s: %%s", destPath, target),
			Quiet:  false,
		}

		writer := &ioprogress.ProgressWriter{
			WriteCloser: f,
			Tracker: &ioprogress.ProgressTracker{
				Handler: func(bytesReceived int64, speed int64) {

					l.Report(fmt.Sprintf("%s (%s/s)\n",
						lxd_shared.GetByteSizeString(bytesReceived, 2),
						lxd_shared.GetByteSizeString(speed, 2)))

					progress.UpdateProgress(ioprogress.ProgressData{
						Text: fmt.Sprintf("%s (%s/s)",
							lxd_shared.GetByteSizeString(bytesReceived, 2),
							lxd_shared.GetByteSizeString(speed, 2))})
				},
			},
		}

		_, err = io.Copy(writer, buf)
		progress.Done("")
		if err != nil {
			l.Report(fmt.Sprintf("Error on pull file %s", target))
			return err
		}

	} else if resp.Type == "symlink" {
		linkTarget, err := ioutil.ReadAll(buf)
		if err != nil {
			return err
		}

		err = os.Symlink(strings.TrimSpace(string(linkTarget)), target)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown file type '%s'", resp.Type)
	}

	return nil
}

func (l *LxdExecutor) ExecCommand(execution *StateExecution, targetHomeDir string, task *tasks.Task) (int, error) {
	instruction := NewInstructionFromTaskWithDebug(*task, "pwd && ls -liah && ")
	env := instruction.EnvironmentMap()

	instruction.Report(l)

	// Set workdir as HOME
	env["HOME"] = targetHomeDir

	// Prepare the command
	req := lxd_api.ContainerExecPost{
		Command:     instruction.ExecutionCommandList(),
		WaitForWS:   true,
		Interactive: false,
		Environment: env,
	}

	execArgs := lxd.ContainerExecArgs{
		// Disable stdin
		Stdin:   ioutil.NopCloser(bytes.NewReader(nil)),
		Stdout:  l.TaskExecutor,
		Stderr:  l.TaskExecutor,
		Control: nil,
		//Control:  handler,
		DataDone: make(chan bool),
	}

	var err error
	// Run the command in the container
	l.CurrentLocalOperation, err = l.LxdClient.ExecContainer(
		execution.Request.ContainerID, req, &execArgs)
	if err != nil {
		l.Report("Error on exec command: " + err.Error())
		execution.UpdateState("error", err, 1)
		return 1, err
	}
	execution.Status = "running"

	// Wait for the operation to complete
	err = l.waitOperation(nil, nil)
	if err != nil {
		l.Report("Error on waiting execution of commands: " + err.Error())
		execution.UpdateState("error", err, 1)
		return 1, err
	}
	opAPI := l.CurrentLocalOperation.Get()
	l.CurrentLocalOperation = nil

	// Wait for any remaining I/O to be flushed
	<-execArgs.DataDone

	// NOTE: If I stop a running container for interrupt execution
	// waitOperation doesn't return error but an empty map as opAPI.
	// I consider it as an error.
	if val, ok := opAPI.Metadata["return"]; ok {
		execution.Result = int(val.(float64))
		l.Report(fmt.Sprintf("========> Execution Exit with value (%d)\n",
			execution.Result))

	} else {
		l.Report(fmt.Sprintf("========> Execution Interrupted (%v)\n",
			opAPI.Metadata))
		execution.Result = 1
		execution.Error = fmt.Errorf("Execution Interrupted")
	}
	if execution.Result == 0 {
		execution.Status = "done"
	} else {
		execution.Status = "error"
	}

	return execution.Result, execution.Error
}

//
func (l *LxdExecutor) recursiveListFile(nameContainer string, targetPath string, list *list.List) error {
	buf, resp, err := l.LxdClient.GetContainerFile(nameContainer, targetPath)
	if err != nil {
		return err
	}
	if buf != nil {
		// Needed to avoid: dial unix /var/lib/lxd/unix.socket: socket: too many open files
		buf.Close()
	}

	if resp.Type == "directory" {
		for _, ent := range resp.Entries {
			nextP := path.Join(targetPath, ent)
			err = l.recursiveListFile(nameContainer, nextP, list)
			if err != nil {
				return err
			}
		}
		list.PushBack(targetPath)
	} else if resp.Type == "file" || resp.Type == "symlink" {
		list.PushFront(targetPath)

	} else {
		l.Report("Find unsupported file type " + resp.Type + ". Skipped.")
	}

	return nil
}

func (l *LxdExecutor) DeleteContainerDirRecursive(containerName, dir string) error {
	var err error
	var list *list.List = list.New()

	// Create list of files/directories to remove. (files are pushed before directories)
	err = l.recursiveListFile(containerName, dir, list)
	if err != nil {
		return err
	}

	for e := list.Front(); e != nil; e = e.Next() {
		l.Report(fmt.Sprintf("Removing old cache file %s...", e.Value.(string)))
		err = l.LxdClient.DeleteContainerFile(containerName, e.Value.(string))
		if err != nil {
			l.Report(fmt.Sprintf("ERROR: Error on removing %s: %s",
				e.Value, err.Error()))
		}
	}

	return nil
}

func (l *LxdExecutor) GetContainerName(task *tasks.Task) string {
	var ans string

	if task.Image != "" {
		image := task.Image
		if len(task.Image) > 20 {
			image = task.Image[:19]
		}

		// To avoid error: Container name isn't a valid hostname
		// I replace any . with -.
		// I can't use / because it's used for snapshots.
		ans = "mottainai-" +
			strings.Replace(strings.Replace(image, "/", "-", -1), ".", "-", -1) +
			"-" + task.ID
	} else {
		ans = "mottainai-" + task.ID
	}

	return ans
}

func (l *LxdExecutor) waitOperation(rawOp interface{}, p *lxd_utils.ProgressRenderer) error {

	var op interface{} = rawOp
	var err error = nil

	// NOTE: currently on ARM we have a weird behavior where the process that waits
	//       for LXD operation often remain blocked. It seems related to a concurrency
	//       problem on initializing Golang channel.
	//       As a workaround, I sleep some seconds before waiting for a response.

	// Retrieve value of sleep before waiting execution of the operation.
	sec, okType := l.Config.GetAgent().LxdCacheRegistry["wait_sleep"]
	if !okType || sec == "" {
		// By default I wait for 1 seconds.
		time.Sleep(1000 * time.Millisecond)
	} else if i, e := strconv.Atoi(sec); e == nil && i > 0 {
		duration, err := time.ParseDuration(fmt.Sprintf("%ds", i))
		if err == nil {
			time.Sleep(duration)
		}
	}

	// TODO: Verify if could be a valid idea permit to use wait not cancelable.
	// err = op.Wait()

	if rawOp == nil {
		if l.CurrentLocalOperation != nil {
			op = l.CurrentLocalOperation
		} else if l.RemoteOperation != nil {
			op = l.RemoteOperation
		} else {
			l.Report("WARN: No operations found.")
		}
	}

	if p != nil {
		err = lxd_utils.CancelableWait(op, p)
	} else {
		err = lxd_utils.CancelableWait(op, nil)
	}

	return err
}
