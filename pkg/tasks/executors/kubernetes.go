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
	"strconv"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesExecutor struct {
	*TaskExecutor
	Namespace        string
	PodID            string
	KubernetesClient *kubernetes.Clientset
}

func NewKubernetesExecutor(config *setting.Config) *KubernetesExecutor {
	return &KubernetesExecutor{
		Namespace: config.GetAgent().KubeNamespace,
		TaskExecutor: &TaskExecutor{
			Context: NewExecutorContext(),
			Config:  config,
		}}
}

func (d *KubernetesExecutor) Setup(docID string) error {
	d.PodID = docID + "-job"
	var config *rest.Config
	var err error
	if d.Config.GetAgent().KubeConfigPath == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}

	} else {
		config, err = clientcmd.BuildConfigFromFlags("", d.Config.GetAgent().KubeConfigPath)
		if err != nil {
			return err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	d.TaskExecutor.Setup(docID)
	d.KubernetesClient = clientset
	return nil
}

func (d *KubernetesExecutor) AttachContainerReport() error {
	p, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Get(d.PodID, metav1.GetOptions{})
	if err != nil {
		return err

	}
	for {
		p, err = d.KubernetesClient.CoreV1().Pods(d.Namespace).Get(d.PodID, metav1.GetOptions{})
		if err != nil {
			return err
		}
		d.Report("Waiting for POD " + d.PodID + " to come up. State: " + string(p.Status.Phase))
		if p.Status.Phase != apiv1.PodPending {
			break
		}
	}

	if p.Status.Phase != apiv1.PodRunning { //something went wrong, but we want to catch the error from the handle, which grabs exit code already
		d.Report("No available output")
		return nil
	}

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

func (d *KubernetesExecutor) Play(docID string) (int, error) {
	task_info, err := tasks.FetchTask(d.MottainaiClient)
	if err != nil {
		return 1, err
	}
	instruction := NewInstructionFromTask(task_info)
	d.Context.ResolveMounts(instruction)

	p, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Create(&apiv1.Pod{

		ObjectMeta: metav1.ObjectMeta{
			Name:      d.PodID,
			Namespace: d.Namespace,
			Labels:    map[string]string{"mottainai-job" + d.PodID: "true"},
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyNever,

			Containers: []apiv1.Container{{
				Name:       d.PodID,
				Command:    instruction.EntrypointList(),
				Args:       instruction.CommandList(),
				Image:      task_info.Image,
				WorkingDir: d.Context.HostPath(task_info.Directory),
				TTY:        true,
			}}}})
	if err != nil {
		return 1, err
	}
	//time.Sleep(10 * time.Second)
	err = d.AttachContainerReport()
	if err != nil {
		return 1, errors.Wrap(err, "Error attaching stdout to pod "+d.PodID)
	}
	defer d.CleanUpContainer()

	return d.Handle(p)
}
func (d *KubernetesExecutor) Clean() error {

	if err := d.TaskExecutor.Clean(); err != nil {
		return err
	}
	return d.CleanUpContainer()
}

func (d *KubernetesExecutor) Handle(p *apiv1.Pod) (int, error) {
	defer d.CleanUpContainer()

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

	// d.Report("Upload of artifacts starts")
	// err := d.UploadArtefacts(mapping.ArtefactPath)
	// if err != nil {
	// 	return 1, err
	// }
	// d.Report("Upload of artifacts terminated")

	return int(p.Status.ContainerStatuses[0].State.Terminated.ExitCode), nil

}

func (d *KubernetesExecutor) CleanUpContainer() error {
	d.Report("Cleanup container")

	err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Delete(d.PodID, &metav1.DeleteOptions{})

	if err != nil {
		d.Report("Container cleanup error: ", err.Error())
		return err
	}

	return nil
}
