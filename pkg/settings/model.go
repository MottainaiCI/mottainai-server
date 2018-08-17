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

package setting

import (
	"encoding/json"
	"reflect"
)

type Setting struct {
	ID    string `json:"id" form:"id"`
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
}

func (s *Setting) IsEnabled() bool {

	if s.Value == "yes" || s.Value == "true" {
		return true
	}
	return false

}
func (s *Setting) IsDisabled() bool {
	return !s.IsEnabled()
}

func (s *Setting) Clear() {
	s.Key = ""
	s.Value = ""
	s.ID = ""
}

// TODO: Port NewUserFromMap Task to same or make it common func
func NewSettingFromMap(t map[string]interface{}) Setting {
	u := &Setting{}
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

func NewSettingFromJson(data []byte) Setting {
	var t Setting
	json.Unmarshal(data, &t)
	return t
}

func (t *Setting) ToMap() map[string]interface{} {

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
