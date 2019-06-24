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
	"bufio"
	"io"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
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
	req := d.KubernetesClient.RESTClient().Get().
		Namespace(d.Namespace).
		Name(d.PodID).
		Resource("pods").
		SubResource("log").
		Param("follow", "true").
		Param("container", d.PodID)

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	defer readCloser.Close()

	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			d.Report(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			d.Report("There was an error with the scanner in attached container:" + err.Error())
			return
		}
	}(readCloser)

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
			Labels:    map[string]string{"mottainai-job": d.PodID},
		},
		Spec: apiv1.PodSpec{Containers: []apiv1.Container{{
			Name:       d.PodID,
			Command:    instruction.EntrypointList(),
			Args:       instruction.CommandList(),
			Image:      task_info.Image,
			WorkingDir: d.Context.HostPath(task_info.Directory),
		}}}})
	if err != nil {
		return 1, err
	}
	err = d.AttachContainerReport()
	if err != nil {
		return 1, err
	}
	defer d.CleanUpContainer()

	return d.Handle(p)
}

func (d *KubernetesExecutor) Handle(p *apiv1.Pod) (int, error) {
	starttime := time.Now()
	var (
		resp *apiv1.Pod = &apiv1.Pod{}
	)

	w, err := d.KubernetesClient.CoreV1().Pods(d.Namespace).Watch(metav1.ListOptions{
		Watch:           true,
		ResourceVersion: p.ResourceVersion,
		FieldSelector:   fields.Set{"metadata.name": d.PodID}.AsSelector().String(),
		LabelSelector:   labels.Everything().String(),
	})
	if err != nil {
		return 1, err
	}

	status := resp.Status

	for {

		c2 := make(chan bool)
		go func() {
			for {
				now := time.Now()
				task_info, err := tasks.FetchTask(d.MottainaiClient)
				if err != nil {
					//fetcher.SetTaskResult("error")
					//fetcher.SetTaskStatus("done")
					d.Report(err.Error())
					return
				}
				timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
				if task_info.IsStopped() || timedout {
					c2 <- true
					return
				}
			}

		}()

		select {
		case events, ok := <-w.ResultChan():
			if !ok {
				break
			}
			resp = events.Object.(*apiv1.Pod)
			status = resp.Status
			if resp.Status.Phase != apiv1.PodPending && status.Phase != apiv1.PodRunning {
				w.Stop()
			}
		case <-c2:
			w.Stop()
			return d.HandleTaskStop(true)
		}
	}

	d.Report("Container execution terminated")

	// d.Report("Upload of artifacts starts")
	// err := d.UploadArtefacts(mapping.ArtefactPath)
	// if err != nil {
	// 	return 1, err
	// }
	// d.Report("Upload of artifacts terminated")

	return int(status.ContainerStatuses[0].State.Terminated.ExitCode), nil

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
