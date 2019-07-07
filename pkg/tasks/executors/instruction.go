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
	"fmt"
	"strings"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

// Instruction represent a set of script + env that has to be executed in a given context
type Instruction interface {
	ToScript() string
	CommandList() []string
	EntrypointList() []string
	ExecutionCommandList() []string

	SetEnvironment(env []string)
	EnvironmentList() []string
	EnvironmentMap() map[string]string
	SetTaskEnvVariables(task_info *tasks.Task, ctx *ExecutorContext)

	SetMounts(mounts []string)
	AddMount(mount string)
	MountsList() []string

	Report(Executor)
}

type DefaultInstruction struct {
	Script      []string
	Environment map[string]string
	Entrypoint  []string
	Mounts      []string
}

func (d *DefaultInstruction) ToScript() string {
	return strings.Join(d.Script, " && ")
}

func (d *DefaultInstruction) CommandList() []string {
	return d.Script
}

func (d *DefaultInstruction) EntrypointList() []string {
	return d.Entrypoint
}

func (d *DefaultInstruction) ExecutionCommandList() []string {
	var ans []string = d.Entrypoint
	ans = append(ans, d.ToScript())
	return ans
}

func (d *DefaultInstruction) AddMount(mount string) {
	d.Mounts = append(d.Mounts, mount)
}

func (d *DefaultInstruction) SetMounts(mounts []string) {
	d.Mounts = mounts
}

func (d *DefaultInstruction) MountsList() []string {
	return d.Mounts
}

func (d *DefaultInstruction) SetEnvironment(env []string) {
	var kv []string
	d.Environment = make(map[string]string)

	for _, v := range env {
		kv = strings.Split(v, "=")
		d.Environment[kv[0]] = kv[1]
	}
}

func (d *DefaultInstruction) EnvironmentList() []string {
	var ans []string
	for k, v := range d.Environment {
		ans = append(ans, fmt.Sprintf("%s=%s", k, v))
	}
	return ans
}

func (d *DefaultInstruction) EnvironmentMap() map[string]string {
	if len(d.Environment) == 0 {
		d.Environment["LC_ALL"] = "en_US.UTF-8"
	}

	return d.Environment
}

func (d *DefaultInstruction) SetTaskEnvVariables(task_info *tasks.Task, ctx *ExecutorContext) {
	if _, ok := d.Environment["MOTTAINAI_TASK_ID"]; !ok {
		d.Environment["MOTTAINAI_TASK_ID"] = task_info.ID
	}

	if _, ok := d.Environment["MOTTAINAI_PIPELINE_ID"]; !ok {
		if task_info.PipelineID != "" {
			d.Environment["MOTTAINAI_PIPELINE_ID"] = task_info.PipelineID
		}
	}
	if _, ok := d.Environment["MOTTAINAI_TASK_NAME"]; !ok {
		if task_info.Name != "" {
			d.Environment["MOTTAINAI_TASK_NAME"] = task_info.Name
		}
	}

	if _, ok := d.Environment["MOTTAINAI_TASK_OWNER_ID"]; !ok {
		d.Environment["MOTTAINAI_TASK_OWNER_ID"] = task_info.Owner
	}

	if _, ok := d.Environment["MOTTAINAI_ARTEFACT_PATH"]; !ok {
		d.Environment["MOTTAINAI_ARTEFACT_PATH"] = ctx.TargetArtefactDir
	}

	if _, ok := d.Environment["MOTTAINAI_STORAGE_PATH"]; !ok {
		d.Environment["MOTTAINAI_STORAGE_PATH"] = ctx.TargetStorageDir
	}

	if _, ok := d.Environment["MOTTAINAI_ROOT_TASK"]; !ok {
		if task_info.RootTask != "" {
			d.Environment["MOTTAINAI_ROOT_TASK"] = task_info.RootTask
		}
	}

	if _, ok := d.Environment["MOTTAINAI_TASK_TYPE"]; !ok {
		d.Environment["MOTTAINAI_TASK_TYPE"] = task_info.Type
	}

	if _, ok := d.Environment["MOTTAINAI_TASK_IMAGE"]; !ok {
		if task_info.Image != "" {
			d.Environment["MOTTAINAI_TASK_IMAGE"] = task_info.Image
		}
	}

	if _, ok := d.Environment["MOTTAINAI_TASK_QUEUE"]; !ok {
		if task_info.Queue != "" {
			d.Environment["MOTTAINAI_TASK_QUEUE"] = task_info.Queue
		}
	}
}

func (instruction *DefaultInstruction) Report(d Executor) {
	d.Report("Entrypoint: ")
	for _, v := range instruction.EntrypointList() {
		d.Report("- " + v)
	}
	d.Report("Commands: ")
	for _, v := range instruction.CommandList() {
		d.Report("- " + v)
	}
	d.Report("Binds: ")
	for _, v := range instruction.MountsList() {
		d.Report("- " + v)
	}
	d.Report("Envs: ")
	for _, v := range instruction.EnvironmentList() {
		// redact env values, display keys
		result := strings.SplitAfter(v, "=")
		if len(result) > 0 {
			d.Report(result[0])
		}
	}
}

func NewDebugInstruction(script []string, debugInstructions string) Instruction {
	return NewBashInstruction([]string{
		debugInstructions + NewDefaultInstruction([]string{}, script).ToScript(),
	})
}

func NewBashInstruction(script []string) Instruction {
	return NewDefaultInstruction([]string{"/bin/bash", "-c"}, script)
}

func NewDefaultInstruction(entrypoint, script []string) Instruction {
	return &DefaultInstruction{Script: script, Entrypoint: entrypoint}
}

func NewInstructionFromTaskWithDebug(task tasks.Task, debugInstructions string) Instruction {
	instruction := NewDebugInstruction(task.Script, debugInstructions)
	if len(task.Entrypoint) > 0 {
		instruction = NewDefaultInstruction(task.Entrypoint, task.Script)
	}
	instruction.SetEnvironment(task.Environment)
	instruction.SetMounts(task.Binds)

	return instruction
}

func NewInstructionFromTask(task tasks.Task) Instruction {
	return NewInstructionFromTaskWithDebug(task, "pwd;ls -liah;")
}
