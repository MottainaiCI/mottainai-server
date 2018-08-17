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

package storage

import (
	"encoding/json"
)

type Storage struct {
	ID string `json:"ID"`
	//Key  string `form:"key" json:"key"`
	Name  string `form:"name" json:"name"`
	Path  string `json:"path" form:"path"`
	Owner string `json:"owner_id" form:"owner_id"`

	//TaskID string `json:"taskid" form:"taskid"`
}

func (t *Storage) IsOwner(id string) bool {

	if id == t.Owner {
		return true
	}

	return false
}
func NewFromJson(data []byte) Storage {
	var t Storage
	json.Unmarshal(data, &t)
	return t
}

func NewFromMap(t map[string]interface{}) Storage {

	var (
		name  string
		path  string
		owner string
	//	key  string
	)

	//if str, ok := t["key"].(string); ok {
	//	key = str
	//}
	if str, ok := t["name"].(string); ok {
		name = str
	}
	if str, ok := t["path"].(string); ok {
		path = str
	}
	if str, ok := t["owner_id"].(string); ok {
		owner = str
	}
	Storage := Storage{
		Name:  name,
		Path:  path,
		Owner: owner,
		//	Key:  key,
	}
	return Storage
}
