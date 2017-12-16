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

package nodes

import "encoding/json"

type Node struct {
	ID     int    `json:"ID"`
	NodeID string `form:"nodeid" json:"nodeid"`
	Key    string `json:"key" form:"key"`
	User   string `json:"user" form:"user"`
	Pass   string `json:"pass" form:"pass"`
	Owner  int    `json:"owner" form:"owner"`
}

func NewFromJson(data []byte) Node {
	var t Node
	json.Unmarshal(data, &t)
	return t
}

func NewNodeFromMap(t map[string]interface{}) Node {

	var (
		key    string
		user   string
		pass   string
		owner  int
		nodeid string
	)

	if str, ok := t["user"].(string); ok {
		user = str
	}
	if str, ok := t["key"].(string); ok {
		key = str
	}
	if str, ok := t["pass"].(string); ok {
		pass = str
	}
	if w, ok := t["owner"].(int); ok {
		owner = w
	}
	if str, ok := t["nodeid"].(string); ok {
		nodeid = str
	}
	node := Node{
		Owner:  owner,
		Pass:   pass,
		Key:    key,
		User:   user,
		NodeID: nodeid,
	}
	return node
}
