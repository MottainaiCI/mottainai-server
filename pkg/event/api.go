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

package event

import (
	"encoding/json"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
)

// APIResponse represent an API payload response
type APIResponse struct {
	ID        string `json:"id,omitempty"`
	ObjType   string `json:"type,omitempty"`
	Processed string `json:"processed,omitempty"`
	Event     string `json:"event,omitempty"`
	Error     string `json:"error,omitempty"`
	Status    string `json:"status,omitempty"`
	Data      string `json:"data,omitempty"`

	Request *schema.Request `json:"-"`
}

// DecodeAPIResponse returns an APIResponse from []byte
func DecodeAPIResponse(data []byte) (APIResponse, error) {
	var r APIResponse
	err := json.Unmarshal(data, &r)
	return r, err
}
