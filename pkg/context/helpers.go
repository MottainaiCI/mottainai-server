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

package context

import (
	"strconv"
	"strings"

	storage "github.com/MottainaiCI/mottainai-server/pkg/storage"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	auth "github.com/MottainaiCI/mottainai-server/pkg/auth"
)

const NameSpacesPrefix = "::"

var noperm = map[string]string{
	"message": "It seems you don't have enough permissions to perform this operation, i'm sorry.",
}

func (c *Context) CheckTaskPermissions(task *task.Task) bool {
	uid, err := strconv.Atoi(task.Owner)
	if err != nil {
		return false
	}

	// Return true if Admin or Owner of it
	if c.User.IsManagerOrAdmin() || c.User.ID == uid {
		return true
	}

	if auth.IsAPIPath(c.Req.URL.Path) {
		c.JSON(403, noperm)
	} else {
		c.Error(403)
	}
	return false
}

func (c *Context) CheckStoragePermissions(storage *storage.Storage) bool {
	uid, err := strconv.Atoi(storage.Owner)
	if err != nil {
		return false
	}

	// Return true if Admin or Owner of it
	if c.User.IsManagerOrAdmin() || c.User.ID == uid {
		return true
	}

	if auth.IsAPIPath(c.Req.URL.Path) {
		c.JSON(403, noperm)
	} else {
		c.Error(403)
	}
	return false
}

// namepath checks
func (c *Context) CheckStorageBelongs(storage string) bool {
	if len(storage) > 0 &&
		!c.User.IsManagerOrAdmin() &&
		!strings.HasPrefix(storage, c.User.Name+NameSpacesPrefix) {

		if auth.IsAPIPath(c.Req.URL.Path) {
			c.JSON(403, noperm)
		} else {
			c.Error(403)
		}
		return false
	}

	return true
}

func (c *Context) CheckNamespaceBelongs(namespace string) bool {
	if len(namespace) > 0 &&
		!c.User.IsManagerOrAdmin() &&
		!strings.HasPrefix(namespace, c.User.Name+NameSpacesPrefix) {

		if auth.IsAPIPath(c.Req.URL.Path) {
			c.JSON(403, noperm)
		} else {
			c.Error(403)
		}
		return false
	}

	return true
}
