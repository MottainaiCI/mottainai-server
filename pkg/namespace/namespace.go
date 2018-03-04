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

import "encoding/json"

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
