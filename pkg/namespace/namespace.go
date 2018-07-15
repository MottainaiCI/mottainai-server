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

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

type Namespace struct {
	ID         int    `json:"ID"`
	Name       string `form:"name" json:"name"`
	Path       string `json:"path" form:"path"`
	Visibility string `json:"visbility" form:"visbility"`
	Owner      string `json:"owner_id" form:"owner_id"`
	//TaskID string `json:"taskid" form:"taskid"`
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

	Namespace := Namespace{
		Name:       name,
		Path:       path,
		Visibility: visibility,
		Owner:      owner,
	}
	return Namespace
}
func (n *Namespace) Exists() bool {

	fi, err := os.Stat(filepath.Join(setting.Configuration.NamespacePath, n.Name))
	if err != nil {
		panic(err)
	}
	if fi.Mode().IsDir() {
		return true
	}
	return false
}

func (n *Namespace) Wipe() {
	os.RemoveAll(filepath.Join(setting.Configuration.NamespacePath, n.Name))
	os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, n.Name), os.ModePerm)
}

func (n *Namespace) Tag(from string) error {

	n.Wipe()

	taskArtefact := filepath.Join(setting.Configuration.ArtefactPath, from)
	namespace := filepath.Join(setting.Configuration.NamespacePath, n.Name)
	return utils.DeepCopy(taskArtefact, namespace)
}

func (n *Namespace) Clone(old Namespace) error {

	n.Wipe()

	oldNamespace := filepath.Join(setting.Configuration.NamespacePath, old.Path)
	newNamespace := filepath.Join(setting.Configuration.NamespacePath, n.Name)
	return utils.DeepCopy(oldNamespace, newNamespace)
}
