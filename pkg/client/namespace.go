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
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

func (f *Fetcher) NamespaceDelete(id string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("delete"),
		Interpolations: map[string]string{
			":name": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) NamespaceRemovePath(id, path string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("remove"),
		Interpolations: map[string]string{
			":id":   id,
			":path": path,
		},
	}

	return f.HandleAPIResponse(req)
}
func (f *Fetcher) NamespaceClone(from, to string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("append"),
		Interpolations: map[string]string{
			":from": from,
			":name": to,
		},
	}

	return f.HandleAPIResponse(req)
}
func (f *Fetcher) NamespaceAppend(id, name string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("append"),
		Interpolations: map[string]string{
			":taskid": id,
			":name":   name,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) NamespaceTag(id, tag string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("tag"),
		Interpolations: map[string]string{
			":taskid": id,
			":name":   tag,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) NamespaceCreate(t string) (event.APIResponse, error) {

	req := Request{
		Route: v1.Schema.GetNamespaceRoute("create"),
		Interpolations: map[string]string{
			":name": t,
		},
	}

	return f.HandleAPIResponse(req)
}
