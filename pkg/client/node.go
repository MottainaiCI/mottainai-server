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

package client

import (
	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

func (d *Fetcher) CreateNode() (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetNodeRoute("create"),
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) RemoveNode(id string) (event.APIResponse, error) {

	req := schema.Request{
		Route:   v1.Schema.GetNodeRoute("delete"),
		Options: map[string]interface{}{":id": id},
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) NodesTask(key string, target interface{}) error {

	req := schema.Request{
		Route:   v1.Schema.GetNodeRoute("show_tasks"),
		Options: map[string]interface{}{":key": key},
		Target:  target,
	}

	err := d.Handle(req)
	if err != nil {
		return err
	}

	return nil
}

func (f *Fetcher) RegisterNode(ID, hostname string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetNodeRoute("register"),
		Options: map[string]interface{}{
			"key":      f.Config.GetAgent().AgentKey,
			"nodeid":   ID,
			"hostname": hostname,
		},
	}

	return f.HandleAPIResponse(req)
}
