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

package namespace

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	cp "github.com/otiai10/copy"
)

type Namespace struct {
	ID         string `json:"ID"`
	Name       string `json:"name" form:"name""`
	Path       string `json:"path" form:"path"`
	Visibility string `json:"visbility" form:"visbility"`
	Owner      string `json:"owner_id" form:"owner_id"`
}

func (u *Namespace) IsPublic() bool {
	if u.Visibility == "public" {
		return true
	}
	return false
}

func (u *Namespace) IsPrivate() bool {
	if u.Visibility == "private" {
		return true
	}
	return false
}

func (u *Namespace) IsOrganization() bool {
	if u.Visibility == "organization" {
		return true
	}
	return false
}

func (u *Namespace) IsGroupVisibile() bool {
	if u.Visibility == "group" {
		return true
	}
	return false
}

func (u *Namespace) IsInternal() bool {
	if u.Visibility == "internal" {
		return true
	}
	return false
}
func (u *Namespace) MakePublic() {
	u.Visibility = "public"
}

func (u *Namespace) MakeInternal() {
	u.Visibility = "internal"
}

func (u *Namespace) MakeGroupVisible() {
	u.Visibility = "group"
}

func (u *Namespace) MakeOrganizationVisible() {
	u.Visibility = "organization"
}

func (u *Namespace) MakePrivate() {
	u.Visibility = "private"
}

func NewFromJson(data []byte) Namespace {
	var t Namespace
	json.Unmarshal(data, &t)
	return t
}

func NewFromMap(t map[string]interface{}) Namespace {

	var (
		name       string
		path       string
		visibility string
		id         string
		owner      string
	)
	if str, ok := t["owner_id"].(string); ok {
		owner = str
	}
	if str, ok := t["name"].(string); ok {
		name = str
	}
	if str, ok := t["path"].(string); ok {
		path = str
	}
	if str, ok := t["visibility"].(string); ok {
		visibility = str
	}
	if str, ok := t["id"].(string); ok {
		id = str
	}
	Namespace := Namespace{
		Name:       name,
		Path:       path,
		Visibility: visibility,
		Owner:      owner,
		ID:         id,
	}
	return Namespace
}
func (n *Namespace) Exists(namespacePath string) bool {

	fi, err := os.Stat(filepath.Join(namespacePath, n.Name))
	if err != nil {
		panic(err)
	}
	if fi.Mode().IsDir() {
		return true
	}
	return false
}

func (n *Namespace) Wipe(namespacePath string) {
	os.RemoveAll(filepath.Join(namespacePath, n.Name))
	os.MkdirAll(filepath.Join(namespacePath, n.Name), os.ModePerm)
}

func (n *Namespace) Tag(
	from string,
	namespacePath string,
	artefactPath string) error {

	n.Wipe(namespacePath)

	taskArtefact := filepath.Join(artefactPath, from)
	namespace := filepath.Join(namespacePath, n.Name)
	return utils.DeepCopy(taskArtefact, namespace)
}

func (n *Namespace) Append(from string,
	namespacePath string,
	artefactPath string) error {

	taskArtefact := filepath.Join(artefactPath, from)
	namespace := filepath.Join(namespacePath, n.Name)

	return utils.DeepCopy(taskArtefact, namespace)
}

func (n *Namespace) Clone(old Namespace, namespacePath string) error {

	n.Wipe(namespacePath)

	oldNamespace := filepath.Join(namespacePath, old.Path)
	newNamespace := filepath.Join(namespacePath, n.Name)

	return cp.Copy(oldNamespace, newNamespace, cp.Options{
		Sync:          true,
		PreserveTimes: true,
		PreserveOwner: true,
	})
}
