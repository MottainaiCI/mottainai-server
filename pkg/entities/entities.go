/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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
package entities

import (
	"errors"
)

type MottainaiEntity string

const (
	Webhooks      MottainaiEntity = "webhooks"
	Tasks         MottainaiEntity = "tasks"
	Secrets       MottainaiEntity = "secrets"
	Users         MottainaiEntity = "users"
	Plans         MottainaiEntity = "plans"
	Pipelines     MottainaiEntity = "pipelines"
	Nodes         MottainaiEntity = "nodes"
	Namespaces    MottainaiEntity = "namespaces"
	Tokens        MottainaiEntity = "tokens"
	Artefacts     MottainaiEntity = "artefacts"
	Storages      MottainaiEntity = "storage"
	Organizations MottainaiEntity = "organizations"
	Settings      MottainaiEntity = "settings"
	Queues        MottainaiEntity = "queues"
	NodeQueues    MottainaiEntity = "nodequeues"
)

func GetMottainaiEntities() []MottainaiEntity {
	return []MottainaiEntity{
		Webhooks,
		Tasks,
		Secrets,
		Users,
		Plans,
		Pipelines,
		Nodes,
		Namespaces,
		Tokens,
		Artefacts,
		Storages,
		Organizations,
		Settings,
		Queues,
		NodeQueues,
	}
}

func NewMottainaiEntity(e string) (MottainaiEntity, error) {
	var ans MottainaiEntity
	var err error = nil

	switch e {
	case "webhooks":
		ans = Webhooks
	case "tasks":
		ans = Tasks
	case "secrets":
		ans = Secrets
	case "users":
		ans = Users
	case "plans":
		ans = Plans
	case "pipelines":
		ans = Pipelines
	case "nodes":
		ans = Nodes
	case "namespaces":
		ans = Namespaces
	case "tokens":
		ans = Tokens
	case "artefacts":
		ans = Artefacts
	case "storages":
		ans = Storages
	case "organizations":
		ans = Organizations
	case "settings":
		ans = Settings
	case "queues":
		ans = Queues
	case "nodequeues":
		ans = NodeQueues
	default:
		err = errors.New("Invalid entity string")
	}
	return ans, err
}

func (e *MottainaiEntity) String() string {
	switch *e {
	case Webhooks:
		return "webhooks"
	case Tasks:
		return "tasks"
	case Secrets:
		return "secrets"
	case Users:
		return "users"
	case Plans:
		return "plans"
	case Pipelines:
		return "pipelines"
	case Nodes:
		return "nodes"
	case Namespaces:
		return "namespaces"
	case Tokens:
		return "tokens"
	case Artefacts:
		return "artefacts"
	case Storages:
		return "storage"
	case Organizations:
		return "organizations"
	case Settings:
		return "settings"
	case Queues:
		return "queues"
	case NodeQueues:
		return "nodequeues"
	}
	return ""
}
