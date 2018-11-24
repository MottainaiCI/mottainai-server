// +build !lxd

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
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func GenDefaultTaskHandler(config *setting.Config) *TaskHandler {
	return &TaskHandler{Tasks: map[string]interface{}{

		"docker_execute": DockerPlayer(config),
		"docker":         DockerPlayer(config),

		"libvirt_execute": LibvirtPlayer(config),
		"libvirt_vagrant": LibvirtPlayer(config),

		"virtualbox_execute": VirtualBoxPlayer(config),
		"virtualbox_vagrant": VirtualBoxPlayer(config),

		"error": HandleErr(config),
		//	"success":        HandleSuccess,
	},
		Config: config,
	}
}
