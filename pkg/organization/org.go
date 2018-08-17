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

package organization

import (
	"encoding/json"
	"reflect"
)

type Organization struct {
	ID   string `json:"id" form:"id"`
	Name string `json:"name" form:"name"`

	Projects []string `json:"projects" form:"projects"`

	Members []string `json:"members" form:"members"`
	Owners  []string `json:"owners" form:"owners"`
	Admins  []string `json:"admins" form:"admins"`
}

func (org *Organization) AddAdmin(s string) {
	org.Admins = append(org.Admins, s)
}
func (org *Organization) AddOwner(s string) {
	org.Owners = append(org.Owners, s)
}
func (org *Organization) AddMember(s string) {
	org.Members = append(org.Members, s)
}
func (org *Organization) ContainsAdmin(s string) bool {
	for _, m := range org.Admins {
		if s == m {
			return true
		}
	}
	return false
}

func (org *Organization) ContainsOwner(s string) bool {
	for _, m := range org.Owners {
		if s == m {
			return true
		}
	}
	return false
}

func (org *Organization) ContainsMember(s string) bool {
	for _, m := range org.Members {
		if s == m {
			return true
		}
	}
	return false
}

// TODO: Port NewUserFromMap Task to same or make it common func
func NewOrganizationFromMap(t map[string]interface{}) Organization {
	u := &Organization{}
	val := reflect.ValueOf(u).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if typeField.Type.String() == "string" {
			if str, ok := t[tag.Get("form")].(string); ok {
				valueField.SetString(str)
			}
		}

		if typeField.Type.String() == "[]string" {
			if b, ok := t[tag.Get("form")].([]string); ok {
				valueField.Set(reflect.ValueOf(b))
			} else if b, ok := t[tag.Get("form")].([]interface{}); ok {
				// convert all to string before set
				var r []string
				for _, f := range b {
					r = append(r, f.(string))
				}
				valueField.Set(reflect.ValueOf(r))
			}
		}
	}
	return *u
}

func NewOrganizationFromJson(data []byte) Organization {
	var t Organization
	json.Unmarshal(data, &t)
	return t
}

func (t *Organization) Clear() {
}

func (t *Organization) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
	}
	return ts
}
