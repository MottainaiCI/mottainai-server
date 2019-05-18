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

func (f *Fetcher) UserCreate(data map[string]interface{}) (event.APIResponse, error) {

	req := schema.Request{
		Route:   v1.Schema.GetUserRoute("create"),
		Options: data,
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) UserRemove(id string) (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetUserRoute("delete"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) UserUpdate(id string, data map[string]interface{}) (event.APIResponse, error) {
	data[":id"] = id

	req := schema.Request{
		Route:   v1.Schema.GetUserRoute("edit"),
		Options: data,
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) UserSet(id, t string) (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetUserRoute("set_" + t),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) UserUnset(id, t string) (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetUserRoute("unset_" + t),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}
