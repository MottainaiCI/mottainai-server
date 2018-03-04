/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package artefact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Artefact struct {
	ID        int    `json:"ID"`
	Name      string `form:"name" json:"name"`
	Path      string `json:"path" form:"path"`
	Task      int    `json:"task" form:"task"`
	Namespace int    `json:"namespace" form:"namespace"`
}

func NewFromJson(data []byte) Artefact {
	var t Artefact
	json.Unmarshal(data, &t)
	return t
}

func (a *Artefact) CleanFromNamespace(namespace string) {
	//NamespacePath
	os.RemoveAll(filepath.Join(setting.Configuration.NamespacePath, namespace, a.Path, a.Name))
}

func (a *Artefact) CleanFromTask() {
	os.RemoveAll(filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(a.Task), a.Path, a.Name))
}

func NewFromMap(t map[string]interface{}) Artefact {

	var (
		name      string
		path      string
		task      int
		namespace int
	)

	if str, ok := t["name"].(string); ok {
		name = str
	}
	if str, ok := t["path"].(string); ok {
		path = str
	}
	if str, ok := t["task"].(int); ok {
		task = str
	}
	if w, ok := t["namespace"].(int); ok {
		namespace = w
	}

	Artefact := Artefact{
		Name:      name,
		Path:      path,
		Task:      task,
		Namespace: namespace,
	}
	return Artefact
}
