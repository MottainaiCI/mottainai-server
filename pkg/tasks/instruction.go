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

import "strings"

// Instruction represent a set of script + env that has to be executed in a given context
type Instruction interface {
	ToScript() string
	CommandList() []string
	EntrypointList() []string

	SetEnvironment(env []string)
	EnvironmentList() []string

	SetMounts(mounts []string)
	AddMount(mount string)
	MountsList() []string

	Report(Executor)
}

type DefaultInstruction struct {
	Script      []string
	Environment []string
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
	d.Environment = env
}

func (d *DefaultInstruction) EnvironmentList() []string {
	return d.Environment
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

func NewDebugInstruction(script []string) Instruction {
	return NewBashInstruction([]string{"pwd;ls -liah;" + NewDefaultInstruction([]string{}, script).ToScript()})
}

func NewBashInstruction(script []string) Instruction {
	cmd := []string{"-c"}
	cmd = append(cmd, script...)
	return NewDefaultInstruction([]string{"/bin/bash"}, cmd)
}

func NewDefaultInstruction(entrypoint, script []string) Instruction {
	return &DefaultInstruction{Script: script, Entrypoint: entrypoint}
}

func NewInstructionFromTask(task Task) Instruction {
	instruction := NewDebugInstruction(task.Script)
	if len(task.Entrypoint) > 0 {
		instruction = NewDefaultInstruction(task.Entrypoint, task.Script)
	}
	instruction.SetEnvironment(task.Environment)
	instruction.SetMounts(task.Binds)

	return instruction
}
