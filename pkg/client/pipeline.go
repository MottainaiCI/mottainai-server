/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>

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

func (f *Fetcher) PipelineDelete(id string) (event.APIResponse, error) {

	req := &schema.Request{
		Route: v1.Schema.GetTaskRoute("pipeline_delete"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) PipelineCreate(taskdata map[string]interface{}) (event.APIResponse, error) {
	req := &schema.Request{
		Route:   v1.Schema.GetTaskRoute("create_pipeline"),
		Options: taskdata,
	}

	return f.HandleAPIResponse(req)
}
