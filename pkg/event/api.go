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

import "encoding/json"

// APIResponse represent an API payload response
type APIResponse struct {
	ID        string `json:"id"`
	ObjType   string `json:"type"`
	Processed string `json:"processed"`
	Event     string `json:"event"`
	Error     string `json:"error"`
	Status    string `json:"status"`
	Data      string `json:"data"`
}

// DecodeAPIResponse returns an APIResponse from []byte
func DecodeAPIResponse(data []byte) (APIResponse, error) {
	var r APIResponse
	err := json.Unmarshal(data, &r)
	return r, err
}
