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

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/koding/vagrantutil"
)

type VagrantExecutor struct {
	*TaskExecutor
	Vagrant  *vagrantutil.Vagrant
	Provider string
}

func NewVagrantExecutor() *VagrantExecutor {
	return &VagrantExecutor{Provider: "libvirt", TaskExecutor: &TaskExecutor{Context: NewExecutorContext()}}
}
func (e *VagrantExecutor) Clean() error {
	e.Prune()
	return e.TaskExecutor.Clean()
}

func (d *VagrantExecutor) Prune() {
	out, err := d.Vagrant.Halt()
	if err != nil {
		d.Report("> Error in halting the machine" + err.Error())
	} else {
		for line := range out {
			d.Report(">" + line.Line)

			if line.Error != nil {
				d.Report(">" + line.Error.Error())
				break
			}
		}
	}
	out, err = d.Vagrant.Destroy()
	if err != nil {
		d.Report("> Error in destroying the machine" + err.Error())
	} else {
		for line := range out {
			d.Report(">" + line.Line)

			if line.Error != nil {
				d.Report(">" + line.Error.Error())
				break
			}
		}
	}
}

func (e *VagrantExecutor) Config(image, rootdir string, t *Task) string {
	var box string
	// TODO: Add CPU and RAM from task
	if utils.IsValidUrl(image) {
		box = `config.vm.box_url = "` + image + `"`
	} else {
		box = `config.vm.box = "` + image + `"`
	}
	git_sourced_repo := e.Context.SourceDir
	artefacts := "artefacts"
	if len(t.ArtefactPath) > 0 {
		artefacts = t.ArtefactPath
	}
	storages := "storage"
	if len(t.StoragePath) > 0 {
		storages = t.StoragePath
	}
	var source string
	if len(git_sourced_repo) > 0 {
		source = `config.vm.synced_folder "` + git_sourced_repo + `", "` + rootdir + `"`
	}
	return `# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
` + box + `
` + source + `
config.vm.synced_folder "` + e.Context.ArtefactDir + `", "` + rootdir + artefacts + `"
config.vm.synced_folder "` + e.Context.StorageDir + `", "` + rootdir + storages + `"

 config.vm.hostname = "vagrant"

  config.vm.provider "virtualbox" do |vb|
    # Use VBoxManage to customize the VM. For example to change memory:
    vb.customize ["modifyvm", :id, "--memory", "2048", "--cpus", "2"]
  end
end
`
}

func (d *VagrantExecutor) Setup(docID string) error {
	d.TaskExecutor.Setup(docID)
	vagrant, err := vagrantutil.NewVagrant(d.Context.BuildDir)
	if err != nil {
		return err
	}
	d.Vagrant = vagrant
	d.Vagrant.ProviderName = d.Provider

	return nil
}

func (d *VagrantExecutor) Play(docID string) (int, error) {
	fetcher := d.MottainaiClient
	th := DefaultTaskHandler()
	task_info := th.FetchTask(fetcher)
	image := task_info.Image
	starttime := time.Now()

	vm_build_dir := "/vagrant/"
	vm_config := d.Config(image, vm_build_dir, &task_info)
	d.Report("Config: " + vm_config)
	d.Vagrant.Create(vm_config)

	var execute_script string
	if err := d.DownloadArtefacts(d.Context.ArtefactDir, d.Context.StorageDir); err != nil {
		return 1, err
	}

	if len(task_info.Script) > 0 {
		execute_script = strings.Join(task_info.Script, " && ")
	}

	if len(task_info.Entrypoint) > 0 {
		execute_script = strings.Join(task_info.Entrypoint, " ") + execute_script
	}

	if len(task_info.Environment) > 0 {
		// FIXME: To do, pass environment to the script
		//containerconfig.Env = task_info.Environment
	}

	d.Report("Image: " + task_info.Image)
	// starts the box
	output, err := d.Vagrant.Up()
	if err != nil {
		d.Report(">" + err.Error())
		return 1, err
	}
	for line := range output {
		now := time.Now()
		task_info = th.FetchTask(fetcher)
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			if timedout {
				d.Report("Task timeout!")
			}
			d.Report(ABORT_EXECUTION_ERROR)
			d.Prune()
			fetcher.AbortTask()
			return 0, errors.New(ABORT_EXECUTION_ERROR)
		}
		d.Report(line.Line)
		if line.Error != nil {
			d.Report(">" + line.Error.Error())
			break
		}
	}

	defer d.Prune()

	out, err := d.Vagrant.SSH(execute_script)
	if err != nil {
		d.Report(out)
		return 1, err
	}

	for res := range out {
		d.Report(res.Line)

		now := time.Now()
		task_info = th.FetchTask(fetcher)
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			if timedout {
				d.Report("Task timeout!")
			}
			d.Report(ABORT_EXECUTION_ERROR)
			d.Prune()
			fetcher.AbortTask()
			return 0, errors.New(ABORT_EXECUTION_ERROR)
		}

		d.Report(out)
		if res.Error != nil {
			err = d.UploadArtefacts(d.Context.ArtefactDir)
			if err != nil {
				return 1, err
			}

			d.Report(res.Error)
			return 1, err
		}
	}

	err = d.UploadArtefacts(d.Context.ArtefactDir)
	if err != nil {
		return 1, err
	}
	// FIXME: exit status change ?  unclear gets caught from Errors already
	return 0, nil
}
