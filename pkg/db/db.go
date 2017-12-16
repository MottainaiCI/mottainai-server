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

package database

type DatabaseInterface interface {
	Init()
	InsertDoc(string, map[string]interface{}) (int, error)
	FindDoc(string, string) (map[int]struct{}, error)
	DeleteDoc(string, int) error
	UpdateDoc(string, int, map[string]interface{}) error
	ReplaceDoc(string, int, map[string]interface{}) error
	GetDoc(string, int) (map[string]interface{}, error)
	AllDocs(string) ([]map[string]interface{}, error)
	DropColl(string) error
}

// For future, now in PoC state will just support
// tiedot
type Database struct {
	Backend string
	DBPath  string
	DBName  string
}

var DBInstance *Database

func NewDatabase(backend string) *Database {
	if DBInstance == nil {
		DBInstance = &Database{Backend: backend}
	}
	DBInstance.Init()
	return DBInstance
}

func Instance() *Database {
	return DBInstance
}
