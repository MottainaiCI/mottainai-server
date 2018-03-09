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
	"strconv"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

type Namespace struct {
	ID   int    `json:"ID"`
	Name string `form:"name" json:"name"`
	Path string `json:"path" form:"path"`
	//TaskID string `json:"taskid" form:"taskid"`
}

func NewFromJson(data []byte) Namespace {
	var t Namespace
	json.Unmarshal(data, &t)
	return t
}

func NewFromMap(t map[string]interface{}) Namespace {

	var (
		name string
		path string
	)

	if str, ok := t["name"].(string); ok {
		name = str
	}
	if str, ok := t["path"].(string); ok {
		path = str
	}

	Namespace := Namespace{
		Name: name,
		Path: path,
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

func (n *Namespace) Tag(from int) error {

	os.RemoveAll(filepath.Join(setting.Configuration.NamespacePath, n.Name))
	os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, n.Name), os.ModePerm)

	source := filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(from))
	return filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		_, file := filepath.Split(path)
		rel := strings.Replace(path, source, "", 1)
		rel = strings.Replace(rel, file, "", 1)

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			return err
		case mode.IsRegular():
			os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, n.Name, rel), os.ModePerm)
			utils.CopyFile(
				path,
				filepath.Join(setting.Configuration.NamespacePath, n.Name, rel, file),
			)
		}
		return nil
	})

}
