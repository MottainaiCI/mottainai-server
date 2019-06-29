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
	"path"

	"strings"
)

type ExecutorContext struct {
	ArtefactDir, StorageDir, NamespaceDir                          string
	BuildDir, SourceDir, RootTaskDir, RealRootDir, TaskRelativeDir string
	DocID                                                          string
	StandardOutput                                                 bool
}

func (ctx *ExecutorContext) ContainerPath(p ...string) string {
	if strings.HasPrefix(p[0], "/") {
		return path.Join(p...)
	}

	var final = path.Join(ctx.RootTaskDir, ctx.TaskRelativeDir)
	for _, a := range p {
		final = path.Join(final, a)
	}
	return final
}

func (ctx *ExecutorContext) HostPath(p ...string) string {
	var final = ctx.RootTaskDir
	for _, a := range p {
		final = path.Join(final, a)
	}
	return final
}

func (ctx *ExecutorContext) ResolveMounts(i Instruction) {
	if len(ctx.SourceDir) > 0 {
		i.AddMount(ctx.SourceDir + ":" + ctx.RootTaskDir)
	}
}

type ArtefactMapping struct {
	ArtefactPath string
	StoragePath  string
}

func (m ArtefactMapping) GetArtefactPath() string {
	if len(m.ArtefactPath) > 0 {
		return m.ArtefactPath
	}

	return "artefacts"
}
func (m ArtefactMapping) GetStoragePath() string {
	if len(m.StoragePath) > 0 {
		return m.StoragePath
	}

	return "storage"
}
func (ctx *ExecutorContext) Report(d Executor) {
	d.Report("Container working dir: " + ctx.HostPath(ctx.TaskRelativeDir))
	d.Report("Context TaskRelativeDir: " + ctx.TaskRelativeDir)
	d.Report("Context BuildDir: " + ctx.BuildDir)
	d.Report("Context SourceDir: " + ctx.SourceDir)
	d.Report("Context RootTaskDir: " + ctx.RootTaskDir)
	d.Report("Context RealRootDir: " + ctx.RealRootDir)
	d.Report("Context StorageDir: " + ctx.StorageDir)
	d.Report("Context ArtefactDir: " + ctx.ArtefactDir)
	d.Report("Context NamespaceDir: " + ctx.NamespaceDir)
}

// ResolveTaskArtefacts given an ArtefactMapping to relative directories and an instruction as input, returns a mapping relative to the host directories used
// and updates the instruction with the new mapping that has been setted
// if hostmapping is setted, it will replicate the same folders between container and host
func (ctx *ExecutorContext) ResolveArtefactsMounts(m ArtefactMapping, i Instruction, hostmapping bool) ArtefactMapping {
	artdir := ctx.ArtefactDir
	storagetmp := ctx.StorageDir

	var storage_path = m.GetStoragePath()
	var artefact_path = m.GetArtefactPath()
	var artefactdir string
	var storagedir string

	if len(m.StoragePath) > 0 {
		storage_path = m.StoragePath
	}

	if hostmapping {
		artefactdir = ctx.HostPath(ctx.TaskRelativeDir, artefact_path)
		storagedir = ctx.HostPath(ctx.TaskRelativeDir, storage_path)
	} else {
		artefactdir = artdir
		storagedir = storagetmp
	}

	i.AddMount(artefactdir + ":" + ctx.ContainerPath(artefact_path))
	i.AddMount(storagedir + ":" + ctx.ContainerPath(storage_path))

	return ArtefactMapping{ArtefactPath: artefactdir, StoragePath: storagedir}
}

func (ctx *ExecutorContext) ResolveArtefacts(m ArtefactMapping) ArtefactMapping {
	var storage_path = "storage"
	var artefact_path = "artefacts"
	var artefactdir string
	var storagedir string

	if len(m.ArtefactPath) > 0 {
		artefact_path = m.ArtefactPath
	}

	if len(m.StoragePath) > 0 {
		storage_path = m.StoragePath
	}

	artefactdir = ctx.ContainerPath(artefact_path)
	storagedir = ctx.ContainerPath(storage_path)

	return ArtefactMapping{ArtefactPath: artefactdir, StoragePath: storagedir}
}

func NewExecutorContext() *ExecutorContext {
	return &ExecutorContext{StandardOutput: true}
}
