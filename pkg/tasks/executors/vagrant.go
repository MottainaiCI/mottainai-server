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
	"os"
	"strings"
	"time"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	vagrantutil "github.com/MottainaiCI/vagrantutil"
)

type VagrantExecutor struct {
	*TaskExecutor
	Vagrant  *vagrantutil.Vagrant
	Provider string
	BoxImage string
}

func NewVagrantExecutor(config *setting.Config) *VagrantExecutor {
	return &VagrantExecutor{
		Provider: "libvirt",
		TaskExecutor: &TaskExecutor{
			Context: NewExecutorContext(),
			Config:  config,
		}}
}

func (e *VagrantExecutor) Clean() error {
	e.Prune()
	return e.TaskExecutor.Clean()
}

func (d *VagrantExecutor) IsLibvirt() bool {
	if d.Provider == "libvirt" {
		return true
	}
	return false
}

func (d *VagrantExecutor) IsVirtualBox() bool {
	if d.Provider == "virtualbox" {
		return true
	}
	return false
}

func (d *VagrantExecutor) BoxRemove(image string) {
	boxes, err := d.Vagrant.BoxList()
	if err != nil {
		d.Report("!! Error in destroying the box " + err.Error())
	}

	for _, box := range boxes {
		d.Report("> Box in machine: " + box.Name)
		if box.Name == image {
			d.Report("> Removing: " + box.Name)
			out, err := d.Vagrant.BoxRemove(box)
			if err != nil {
				d.Report("!! Error in destroying the box " + err.Error())
			} else {
				d.reportOutput(out)
			}
		}
	}
	if d.IsLibvirt() {
		// Sadly vagrant doesn't remove them from pools
		cmdName := "virsh"

		args := []string{"vol-delete", "--pool", "default", image + "_vagrant_box_image_0.img"}
		out, stderr, err := utils.Cmd(cmdName, args)
		if err != nil {
			d.Report("!! There was an error running virsh command: ", err.Error()+": "+stderr)
		}
		d.Report(out)

		args = []string{"pool-destroy", "default"}
		out, stderr, err = utils.Cmd(cmdName, args)
		if err != nil {
			d.Report("!! There was an error running virsh command: ", err.Error()+": "+stderr)
		}
		d.Report(out)

		args = []string{"pool-undefine", "default"}
		out, stderr, err = utils.Cmd(cmdName, args)
		if err != nil {
			d.Report("!! There was an error running virsh command: ", err.Error()+": "+stderr)
		}
		d.Report(out)
	}
}

func (d *VagrantExecutor) Prune() {

	out, err := d.Vagrant.PowerOff()
	if err != nil {
		d.Report("!! Error in halting the machine" + err.Error())
	} else {
		d.reportOutput(out)
	}

	d.BoxRemove(d.BoxImage)
	out, err = d.Vagrant.Destroy()
	if err != nil {
		d.Report("!! Error in destroying the machine" + err.Error())
	} else {
		d.reportOutput(out)
	}

}

func (e *VagrantExecutor) reportOutput(out <-chan *vagrantutil.CommandOutput) {
	for res := range out {
		e.Report(">" + res.Line)
		if res.Error != nil {
			e.Report("!! " + res.Error.Error())
			return
		}
	}
}

func (e *VagrantExecutor) Config(image, rootdir string, t *tasks.Task) string {
	var box, box_url string
	// TODO: Add CPU and RAM from task
	if utils.IsValidUrl(image) {
		box_url = `config.vm.box_url = "` + image + `"`
		box = `config.vm.box = "` + t.ID + `"`
		e.BoxImage = t.ID
	} else {
		box = `config.vm.box = "` + image + `"`
		e.BoxImage = image
	}
	artefacts := "artefacts"
	if len(t.ArtefactPath) > 0 {
		artefacts = t.ArtefactPath
	}
	storages := "storage"
	if len(t.StoragePath) > 0 {
		storages = t.StoragePath
	}
	var env string

	ram := "2048"
	cpu := "1"
	for _, s := range t.Environment {
		env = env + " export " + s + "\n"
		if strings.Contains(s, "CPU") {
			cpu = strings.Replace(s, "CPU=", "", -1)
		}
		if strings.Contains(s, "RAM") {
			ram = strings.Replace(s, "RAM=", "", -1)
		}
	}

	return `# -*- mode: ruby -*-
# vi: set ft=ruby :
$set_environment_variables = <<SCRIPT
tee "/etc/profile.d/myvars.sh" > "/dev/null" <<EOF
` + env + `
EOF
SCRIPT
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
 ` + box + `
 ` + box_url + `
 config.vm.synced_folder "` + e.Context.ArtefactDir + `", "` + rootdir + artefacts + `"
 config.vm.synced_folder "` + e.Context.StorageDir + `", "` + rootdir + storages + `"

 config.vm.hostname = "vagrant"
 config.vm.provision "shell", inline: $set_environment_variables, run: "always"

 config.vm.provider :libvirt do |libvirt|
	 libvirt.storage_pool_name = "default"
	 libvirt.cpus = ` + cpu + `
	 libvirt.memory = ` + ram + `
 end

  config.vm.provider "virtualbox" do |vb|
    vb.customize ["modifyvm", :id, "--memory", "` + ram + `", "--cpus", "` + cpu + `"]
  end
end
`
}

func (d *VagrantExecutor) Setup(docID string) error {
	d.TaskExecutor.Setup(docID)
	var box string
	if len(d.Context.SourceDir) > 0 {
		box = d.Context.SourceDir
	} else {
		box = d.Context.BuildDir
	}
	vagrant, err := vagrantutil.NewVagrant(box)
	if err != nil {
		return err
	}
	d.Vagrant = vagrant
	d.Vagrant.ProviderName = d.Provider

	if d.IsVirtualBox() {
		// Confine VirtualBox environment to BuildDir
		os.Setenv("VAGRANT_HOME", d.TaskExecutor.Config.GetAgent().BuildPath)
		os.Setenv("VBOX_USER_HOME", d.TaskExecutor.Config.GetAgent().BuildPath)
		os.Unsetenv("HOME")
		os.Setenv("HOME", d.TaskExecutor.Config.GetAgent().BuildPath)
	}

	if d.IsLibvirt() {
		cmdName := "virsh"

		args := []string{"pool-define-as", "default", "--type", "dir", "--target", "/var/lib/libvirt/images"}
		out, stderr, err := utils.Cmd(cmdName, args)
		if err != nil {
			d.Report("!! There was an error running virsh command: ", err.Error()+": "+stderr)
		}
		d.Report(out)

		args = []string{"pool-start", "default"}
		out, stderr, err = utils.Cmd(cmdName, args)
		if err != nil {
			d.Report("!! There was an error running virsh command: ", err.Error()+": "+stderr)
		}
		d.Report(out)
	}
	return nil
}

func (d *VagrantExecutor) Play(docID string) (int, error) {
	fetcher := d.MottainaiClient
	task_info, err := tasks.FetchTask(fetcher)
	if err != nil {
		return 1, err
	}
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

	d.Report("Image: " + task_info.Image)
	// starts the box
	output, err := d.Vagrant.Up()
	if err != nil {
		d.Report(">" + err.Error())
		return 1, err
	}
	for line := range output {
		now := time.Now()
		task_info, err = tasks.FetchTask(fetcher)
		if err != nil {
			return 1, err
		}
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			return d.HandleTaskStop(timedout)
		}
		d.Report(line.Line)
		if line.Error != nil {
			d.Report("!! " + line.Error.Error())
			return 1, line.Error
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
		task_info, err = tasks.FetchTask(fetcher)
		if err != nil {
			return 1, err
		}
		timedout := (task_info.TimeOut != 0 && (now.Sub(starttime).Seconds() > task_info.TimeOut))
		if task_info.IsStopped() || timedout {
			return d.HandleTaskStop(timedout)
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
