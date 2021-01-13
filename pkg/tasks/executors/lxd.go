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
	"errors"
	"fmt"
	"strconv"
	"time"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"container/list"
	"os"
	"path"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	lxd_compose "github.com/MottainaiCI/lxd-compose/pkg/executor"

	lxd "github.com/lxc/lxd/client"
	lxd_api "github.com/lxc/lxd/shared/api"
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
	Executor *lxd_compose.LxdCExecutor
}

func (e *LxdExecutor) Prune() {
	// TODO
}

func (l *LxdExecutor) Setup(docID string) error {
	l.TaskExecutor.Setup(docID)

	var err error
	var configPath string = path.Join(l.Config.GetAgent().BuildPath, "/lxc/config.yml")

	if len(l.Config.GetAgent().LxdConfigDir) > 0 {
		// TODO: handle path
		configPath = path.Join(l.Config.GetAgent().LxdConfigDir, "/config.yml")
	}

	waitSleep := 1

	sec, okType := l.Config.GetAgent().LxdCacheRegistry["wait_sleep"]
	if okType {
		if i, e := strconv.Atoi(sec); e == nil && i > 0 {
			waitSleep = i
		}
	}

	l.Executor = lxd_compose.NewLxdCExecutorWithEmitter(
		l.Config.GetAgent().LxdEndpoint,
		configPath,
		[]string{},
		l.Config.GetAgent().LxdEphemeralContainers,
		true, // We want commands output
		true, // We want runtime commands output
		l,    // LxdExecutor implements LxdCExecutorEmitter
	)

	l.Executor.WaitSleep = waitSleep

	err = l.Executor.Setup()
	if err != nil {
		return (errors.New("Error on setup LXD Executor: " + err.Error()))
	}

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
	err = l.Executor.RecursivePushFile(containerName, localWorkDir, l.Context.RootTaskDir)
	if err != nil {
		return "", err
	}
	// Create artefactdir on container
	err = l.Executor.RecursivePushFile(containerName,
		strings.TrimRight(l.Context.ArtefactDir, "/")+"/", mapping.ArtefactPath)
	if err != nil {
		return "", err
	}
	// Create storagedir on container
	err = l.Executor.RecursivePushFile(containerName,
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
	ephemeral := l.Config.GetAgent().LxdEphemeralContainers
	if cachedImage {
		ephemeral = false
	}

	l.Report(">> Creating container " + containerName + "...")
	err = l.Executor.LaunchContainerType(
		containerName, imageFingerprint,
		l.Config.GetAgent().LxdProfiles,
		ephemeral,
	)
	if err != nil {
		l.Report("Creating container error: " + err.Error())
		return 1, err
	}
	defer l.Executor.DeleteContainer(containerName)

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
			err = l.Executor.DoAction2Container(exec.Request.ContainerID, "stop")
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
	err = l.Executor.RecursivePullFile(exec.Request.ContainerID, mapping.ArtefactPath,
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

	if remote == l.Executor.LxdConfig.DefaultRemote {
		l.Report("Remote server is equal to local. Nothing to push.")
		return nil
	}

	image_server, err = l.Executor.LxdConfig.GetInstanceServer(remote)
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

	err = l.Executor.CopyImage(fingerprint, l.Executor.LxdClient, image_server)
	if err != nil {
		return fmt.Errorf("Error on copy image %s: %s", fingerprint, err)
	}

	return nil
}

// Create image on local LXD instance from an active container.
func (l *LxdExecutor) CommitImage(containerName, aliasName string, task *tasks.Task) (string, error) {
	var description string

	// Initialize properties for images
	properties := map[string]string{}
	aliases := []string{aliasName}

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

	compression, _ := l.Config.GetAgent().LxdCacheRegistry["compression_algorithm"]

	return l.Executor.CreateImageFromContainer(containerName, aliases, properties, compression, true)
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
		err = l.Executor.DoAction2Container(containerName, "stop")
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

func (l *LxdExecutor) FindImage(image string) (string, lxd.ImageServer, string, error) {
	var err error
	var tmp_srv, srv lxd.ImageServer
	var img, tmp_img *lxd_api.Image
	var fingerprint string = ""
	var srv_name string = ""

	for remote, server := range l.Executor.LxdConfig.Remotes {
		tmp_srv, err = l.Executor.LxdConfig.GetImageServer(remote)
		if err != nil {
			err = nil
			l.Report(fmt.Sprintf(
				"Error on retrieve ImageServer for remote %s at addr %s",
				remote, server.Addr,
			))
			continue
		}
		tmp_img, err = l.Executor.GetImage(image, tmp_srv)
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
	image, _, _ := l.Executor.LxdClient.GetImage(imageFingerprint)
	if image == nil {
		// NOTE: In concurrency could be happens that different image that
		//       share same aliases generate reset of aliases but
		//       if I work with fingerprint after FindImage I can ignore
		//       aliases.

		// Delete local image with same target aliases to avoid error on pull.
		err = l.Executor.DeleteImageAliases4Alias(imageAlias, l.Executor.LxdClient)

		// Try to pull image to lxd instance
		l.Report(fmt.Sprintf(
			"Try to download image %s from remote %s...",
			imageFingerprint, remote_name,
		))
		err = l.Executor.DownloadImage(imageFingerprint, remote)
	} else {
		l.Report("Image " + imageFingerprint + " already present.")
	}

	return imageFingerprint, err
}

func (l *LxdExecutor) ExecCommand(execution *StateExecution, targetHomeDir string, task *tasks.Task) (int, error) {
	instruction := NewInstructionFromTaskWithDebug(*task, "pwd && ls -liah && ")
	instruction.SetTaskEnvVariables(task, l.Context)
	env := instruction.EnvironmentMap()

	instruction.Report(l)

	// Set workdir as HOME
	env["HOME"] = targetHomeDir

	res, err := l.Executor.RunCommandWithOutput(
		execution.Request.ContainerID, instruction.ToScript(), env,
		l, // output writecloser
		l, // err write closer
		instruction.EntrypointList(),
	)

	// NOTE: If I stop a running container for interrupt execution
	// waitOperation doesn't return error but an empty map as opAPI.
	// I consider it as an error.
	if err == nil {
		execution.Result = res
		l.Report(fmt.Sprintf("========> Execution Exit with value (%d)\n",
			execution.Result))

	} else {
		l.Report(fmt.Sprintf("========> Execution Interrupted (%s)\n",
			err.Error()))
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
	buf, resp, err := l.Executor.LxdClient.GetContainerFile(nameContainer, targetPath)
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
		err = l.Executor.LxdClient.DeleteContainerFile(containerName, e.Value.(string))
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
