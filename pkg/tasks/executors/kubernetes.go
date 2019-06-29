/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>
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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cp "k8s.io/kubernetes/pkg/kubectl/cmd/cp"
)

type KubernetesExecutor struct {
	*TaskExecutor
	Namespace, StorageClass                       string
	PodID, ArtefactPVCID, StoragePVCID, RepoPVCID string
	KubernetesClient                              *kubernetes.Clientset
	KubernetesConfig                              *rest.Config
	UsedPODs, UsedPVCs                            []string
}

func NewKubernetesExecutor(config *setting.Config) *KubernetesExecutor {
	return &KubernetesExecutor{
		Namespace:    config.GetAgent().KubeNamespace,
		StorageClass: config.GetAgent().KubeStorageClass,

		TaskExecutor: &TaskExecutor{
			Context: NewExecutorContext(),
			Config:  config,
		}}
}

func (d *KubernetesExecutor) Setup(docID string) error {
	d.PodID = docID + "-job"
	d.ArtefactPVCID = d.PodID + "-pvc-artefacts"
	d.StoragePVCID = d.PodID + "-pvc-storage"
	d.RepoPVCID = d.PodID + "-pvc-repodata"

	var err error
	if d.Config.GetAgent().KubeConfigPath == "" {
		d.KubernetesConfig, err = rest.InClusterConfig()
		if err != nil {
			return err
		}

	} else {
		d.KubernetesConfig, err = clientcmd.BuildConfigFromFlags("", d.Config.GetAgent().KubeConfigPath)
		if err != nil {
			return err
		}
	}

	clientset, err := kubernetes.NewForConfig(d.KubernetesConfig)
	if err != nil {
		return err
	}

	d.TaskExecutor.Setup(docID)
	d.KubernetesClient = clientset
	return nil
}

func (d *KubernetesExecutor) WaitUntilRunning(pod, namespace string) error {
	p, err := d.KubernetesClient.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	if err != nil {
		return err

	}
	for {
		p, err = d.KubernetesClient.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
		if err != nil {
			return err
		}
		d.Report("Waiting for POD " + pod + " to come up. State: " + string(p.Status.Phase))
		if p.Status.Phase != apiv1.PodPending {
			break
		}
	}

	if p.Status.Phase != apiv1.PodRunning { //something went wrong, but we want to catch the error from the handle, which grabs exit code already
		d.Report("No available output")
		return errors.New("POD failed to start")
	}
	return nil
}

func (d *KubernetesExecutor) AttachContainerReport() error {

	d.WaitUntilRunning(d.PodID, d.Namespace)
	d.Report("Attaching to POD output..")

	req := d.KubernetesClient.CoreV1().RESTClient().Get().
		Namespace(d.Namespace).
		Name(d.PodID).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true)).
		Param("container", d.PodID).
		Param("previous", strconv.FormatBool(false)).
		Param("timestamps", strconv.FormatBool(true))
	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	go func() {
		defer readCloser.Close()
		_, err = io.Copy(d, readCloser)
		if err != nil {
			d.Report("Error: " + err.Error())
			return
		}
	}()

	return nil
}

func getVolume(name, path string) (apiv1.Volume, apiv1.VolumeMount) {
	mount := apiv1.VolumeMount{
		Name:      name,
		MountPath: path,
	}

	vol := apiv1.Volume{
		Name: name,
		VolumeSource: apiv1.VolumeSource{
			PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
			},
		},
	}

	return vol, mount
}

func (d *KubernetesExecutor) CreatePVC(id, size string) error {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return errors.Wrap(err, "invalid quantity")
	}

	_, err = d.KubernetesClient.CoreV1().PersistentVolumeClaims(d.Namespace).Create(&apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   id,
			Labels: map[string]string{},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			StorageClassName: &d.StorageClass,
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					"storage": quantity,
				},
			},
		},
	})
	return err
}

func (d *KubernetesExecutor) DeletePVC(id string) error {
	err := d.KubernetesClient.CoreV1().PersistentVolumeClaims(d.Namespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting persistent volume claim "+id)
	}
	return nil
}

func (d *KubernetesExecutor) PopulateArtefacts(volumes []apiv1.Volume, volumeMounts []apiv1.VolumeMount, srcMapping, outMapping ArtefactMapping) error {

	//  Artefact Uploader
	d.Report("Populating volume mounts")

	stagerPod := d.PodID + "-stager"
	d.UsedPODs = append(d.UsedPODs, stagerPod)

	if _, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stagerPod,
			Namespace: d.Namespace,
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,
			Volumes:       volumes,
			Containers: []apiv1.Container{{
				Name:         d.PodID,
				Command:      []string{"/bin/sh", "-c"},
				Args:         []string{"/bin/tail", "-f", "/dev/null"},
				Image:        d.Config.GetAgent().KubeDropletImage,
				VolumeMounts: volumeMounts,
				TTY:          true,
			},
			}}}); err != nil {
		return err
	}

	if err := d.WaitUntilRunning(stagerPod, d.Namespace); err != nil {
		return err
	}
	if _, err := os.Stat(outMapping.ArtefactPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(outMapping.ArtefactPath)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := d.KubeCP(path.Join(outMapping.ArtefactPath, f.Name()), d.Namespace+"/"+stagerPod+":"+d.Context.ContainerPath(srcMapping.GetArtefactPath())); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(outMapping.StoragePath); !os.IsNotExist(err) {

		files, err := ioutil.ReadDir(outMapping.StoragePath)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := d.KubeCP(path.Join(outMapping.StoragePath, f.Name()), d.Namespace+"/"+stagerPod+":"+d.Context.ContainerPath(srcMapping.GetStoragePath())); err != nil {
				return err
			}
		}

	}
	if len(d.Context.SourceDir) > 0 {
		if _, err := os.Stat(d.Context.SourceDir); !os.IsNotExist(err) {

			files, err := ioutil.ReadDir(d.Context.SourceDir)
			if err != nil {
				return err
			}
			for _, f := range files {
				if err := d.KubeCP(path.Join(d.Context.SourceDir, f.Name()), d.Namespace+"/"+stagerPod+":"+d.Context.RootTaskDir); err != nil {
					return err
				}
			}
		}
	}

	d.Report("Droplet populated from artefacts")

	if err := d.DeletePOD(stagerPod, d.Namespace); err != nil {
		return err
	}

	// END Uploader Artefacts
	return nil
}

func (d *KubernetesExecutor) Play(docID string) (int, error) {
	task_info, err := tasks.FetchTask(d.MottainaiClient)
	if err != nil {
		return 1, err
	}

	instruction := NewInstructionFromTask(task_info)
	d.Context.ResolveMounts(instruction)
	srcMapping := ArtefactMapping{
		ArtefactPath: task_info.ArtefactPath,
		StoragePath:  task_info.StoragePath,
	}
	outMapping := d.Context.ResolveArtefactsMounts(srcMapping, instruction, d.Config.GetAgent().DockerInDocker)

	instruction.Report(d)
	d.Context.Report(d)

	if err := d.DownloadArtefacts(outMapping.ArtefactPath, outMapping.StoragePath); err != nil {
		return 1, err
	}

	for _, s := range []string{d.ArtefactPVCID, d.StoragePVCID, d.RepoPVCID} {
		if task_info.Quota != "" {
			err := d.CreatePVC(s, task_info.Quota)
			if err != nil {
				return 1, err
			}
		} else {
			err := d.CreatePVC(s, d.Config.GetAgent().DefaultTaskQuota)
			if err != nil {
				return 1, err
			}
		}
		d.UsedPVCs = append(d.UsedPVCs, s)
	}

	artefactsVolume, artefactsVolumeMount := getVolume(d.ArtefactPVCID, d.Context.ContainerPath(srcMapping.GetArtefactPath()))
	storageVolume, storageVolumeMount := getVolume(d.StoragePVCID, d.Context.ContainerPath(srcMapping.GetStoragePath()))
	repoVolume, repoVolumeMount := getVolume(d.RepoPVCID, d.Context.RootTaskDir)

	volumes := []apiv1.Volume{artefactsVolume, storageVolume, repoVolume}
	volumesMounts := []apiv1.VolumeMount{artefactsVolumeMount, storageVolumeMount, repoVolumeMount}

	if err := d.PopulateArtefacts(volumes, volumesMounts, srcMapping, outMapping); err != nil {
		return 1, err
	}

	//ctx.SourceDir + ":" + ctx.RootTaskDir
	p, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Create(&apiv1.Pod{

		ObjectMeta: metav1.ObjectMeta{
			Name:      d.PodID,
			Namespace: d.Namespace,
			Labels:    map[string]string{"mottainai-job" + d.PodID: "true"},
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,
			Volumes:       volumes,
			Containers: []apiv1.Container{{
				Name:         d.PodID,
				Command:      instruction.EntrypointList(),
				Args:         instruction.CommandList(),
				Image:        task_info.Image,
				WorkingDir:   d.Context.HostPath(task_info.Directory),
				TTY:          true,
				VolumeMounts: volumesMounts,
			},
			}}})
	if err != nil {
		return 1, err
	}
	d.UsedPODs = append(d.UsedPODs, d.PodID)
	//time.Sleep(10 * time.Second)
	err = d.AttachContainerReport()
	if err != nil {
		return 1, errors.Wrap(err, "Error attaching stdout to pod "+d.PodID)
	}

	return d.Handle(p, instruction, srcMapping, outMapping)
}
func (d *KubernetesExecutor) Clean() error {
	if err := d.TaskExecutor.Clean(); err != nil {
		return err
	}
	return d.CleanUpContainer()
}

func (d *KubernetesExecutor) Handle(p *apiv1.Pod, i Instruction, srcMapping, outMapping ArtefactMapping) (int, error) {

	starttime := time.Now()
	p, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Get(d.PodID, metav1.GetOptions{})
	if err != nil {
		return 1, err
	}

	for {
		p, err = d.KubernetesClient.CoreV1().Pods(d.Namespace).Get(d.PodID, metav1.GetOptions{})
		if err != nil {

			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			d.Report(err.Error())
			return 1, err

		}
		time.Sleep(2 * time.Second)
		now := time.Now()
		task_info, err := tasks.FetchTask(d.MottainaiClient)
		if err != nil {

			//fetcher.SetTaskResult("error")
			//fetcher.SetTaskStatus("done")
			d.Report(err.Error())
			return 1, err

		}
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			return d.HandleTaskStop(true)
		}

		if p.Status.Phase != apiv1.PodPending && p.Status.Phase != apiv1.PodRunning {
			break
		}

	}

	d.Report("Container execution terminated")

	//  Artefact Uploader
	d.Report("Creating uploader droplet")

	uploaderPodName := d.PodID + "-uploader"
	artefactsVolume, artefactsVolumeMount := getVolume(d.ArtefactPVCID, d.Context.ContainerPath(srcMapping.GetArtefactPath()))
	d.UsedPODs = append(d.UsedPODs, uploaderPodName)

	_, err = d.KubernetesClient.CoreV1().Pods(d.Namespace).Create(&apiv1.Pod{

		ObjectMeta: metav1.ObjectMeta{
			Name:      uploaderPodName,
			Namespace: d.Namespace,
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,
			Volumes: []apiv1.Volume{
				artefactsVolume,
			},
			Containers: []apiv1.Container{{
				Name:    d.PodID,
				Command: []string{"/bin/sh", "-c"},
				Args:    []string{"/bin/tail", "-f", "/dev/null"},
				Image:   d.Config.GetAgent().KubeDropletImage,
				VolumeMounts: []apiv1.VolumeMount{
					artefactsVolumeMount,
				},
				TTY: true,
			},
			}}})
	if err != nil {
		return 1, err
	}

	d.WaitUntilRunning(uploaderPodName, d.Namespace)
	if err != nil {
		return 1, err
	}

	if err := d.KubeCP(d.Namespace+"/"+uploaderPodName+":"+d.Context.ContainerPath(srcMapping.GetArtefactPath()), outMapping.ArtefactPath); err != nil {
		return 1, err
	}

	// d.Report("Upload of artifacts starts")
	err = d.UploadArtefacts(outMapping.ArtefactPath)
	if err != nil {
		return 1, err
	}
	d.Report("Upload of artifacts terminated")

	// END Uploader Artefacts

	return int(p.Status.ContainerStatuses[0].State.Terminated.ExitCode), nil

}

func (d *KubernetesExecutor) KubeCP(src, dst string) error {
	copy := cp.NewCopyOptions(genericclioptions.IOStreams{
		In:     nil,
		Out:    d,
		ErrOut: d,
	})

	copy.NoPreserve = true
	copy.Clientset = d.KubernetesClient
	copy.ClientConfig = d.KubernetesConfig
	// FIXME: the code below should work, but doesn't. fallback to kubectl and require it installed in the agent
	// d.Report([]string{d.Namespace + "/" + uploaderPodName + ":" + d.Context.ContainerPath(srcMapping.GetArtefactPath()), outMapping.ArtefactPath})
	// err = copy.Run([]string{d.Namespace + "/" + uploaderPodName + ":" + d.Context.ContainerPath(srcMapping.GetArtefactPath()), outMapping.ArtefactPath})
	// if err != nil {
	// 	return 1, err
	// }
	d.Report("Copying " + src + " to " + dst)
	cmd := exec.Command("kubectl", "cp", src, dst)
	cmd.Env = append(os.Environ(),
		"KUBECONFIG="+d.Config.GetAgent().KubeConfigPath,
	)
	cmd.Stdout = d
	cmd.Stderr = d
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "Failed copying artefacts")
	}
	return nil
}

func (d *KubernetesExecutor) DeletePOD(pod, Namespace string) error {
	d.Report("Cleanup POD: " + pod)

	err := d.KubernetesClient.CoreV1().Pods(Namespace).Delete(pod, &metav1.DeleteOptions{})

	if err != nil {
		d.Report("error deleting pod: ", err.Error())
		return errors.Wrap(err, "Error deleting POD")
	}

	return nil
}

func (d *KubernetesExecutor) CleanUpContainer() error {
	d.Report("Cleanup container")
	// err := d.DeletePOD(d.PodID, d.Namespace)
	// if err != nil {
	// 	d.Report("Container cleanup error: ", err.Error())
	// 	return errors.Wrap(err, "Error cleaning up")
	// }

	for _, gcPOD := range d.UsedPODs {
		err := d.DeletePOD(gcPOD, d.Namespace)
		if err != nil {
			d.Report("Error deleting POD " + gcPOD + ": " + err.Error())
		}
	}
	for _, gcPVC := range d.UsedPVCs {
		err := d.DeletePVC(gcPVC)
		if err != nil {
			d.Report("Error deleting PVC " + gcPVC + ": " + err.Error())
		}
	}

	return nil
}
