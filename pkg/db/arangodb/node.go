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

package arangodb

import (
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
)

var NodeColl = "Nodes"

func (d *Database) IndexNode() {
	d.AddIndex(NodeColl, []string{"nodeid"})
	d.AddIndex(NodeColl, []string{"key"})
}

func (d *Database) CreateNode(t map[string]interface{}) (string, error) {
	return d.InsertDoc(NodeColl, t)
}
func (d *Database) InsertNode(n *nodes.Node) (string, error) {
	return d.CreateNode(n.ToMap())
}
func (d *Database) DeleteNode(docID string) error {
	return d.DeleteDoc(NodeColl, docID)
}

func (d *Database) UpdateNode(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(NodeColl, docID, t)
}

func (d *Database) GetNode(docID string) (nodes.Node, error) {
	doc, err := d.GetDoc(NodeColl, docID)
	if err != nil {
		return nodes.Node{}, err
	}
	t := nodes.NewNodeFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) GetNodeByKey(key string) (nodes.Node, error) {
	var res []nodes.Node

	queryResult, err := d.FindDoc("", `FOR c IN `+NodeColl+`
    FILTER c.key == "`+key+`"
    RETURN c`)
	if err != nil || len(queryResult) != 1 {
		return nodes.Node{}, nil
	}

	// Query result are document IDs
	for id, doc := range queryResult {
		n := nodes.NewNodeFromMap(doc.(map[string]interface{}))
		n.ID = id
		res = append(res, n)
	}
	return res[0], nil

}

func (d *Database) ListNodes() []dbcommon.DocItem {
	return d.ListDocs(NodeColl)
}

func (d *Database) AllNodes() []nodes.Node {
	node_list := make([]nodes.Node, 0)

	docs, err := d.FindDoc("", "FOR c IN "+NodeColl+" return c")
	if err != nil {
		return []nodes.Node{}
	}

	for id, doc := range docs {
		n := nodes.NewNodeFromMap(doc.(map[string]interface{}))
		n.ID = id
		node_list = append(node_list, n)
	}

	return node_list
}
