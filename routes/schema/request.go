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
package schema

import (
	"io"
	"net/http"
)

type Request struct {
	Route   Route
	Options map[string]interface{}
	Target  interface{}
	Body    io.Reader
}

func (req *Request) NewAPIHTTPRequest(endpoint string) (*http.Request, error) {
	interpolations := req.Options
	relaxedInterpolations := req.Route.RemoveInterpolations(req.Options)
	req.Options = relaxedInterpolations

	return req.Route.NewAPIRequest(endpoint, interpolations, req.Body)
}
