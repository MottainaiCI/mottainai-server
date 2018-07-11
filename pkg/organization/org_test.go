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

package organization

import "testing"

func TestNewOrgFromMap(t *testing.T) {

	U := NewOrganizationFromMap(map[string]interface{}{
		"owners":   []interface{}{"mudler", "borg"},
		"projects": []string{"foo", "bar"},
		"members":  []string{"1", "2"},
		"name":     "42",
	})
	U.AddAdmin("duck")
	U.AddAdmin("go")
	if !U.ContainsOwner("mudler") {
		t.Error("Invalid owners", U)
		return
	}

	if U.Owners[0] != "mudler" || U.Owners[1] != "borg" ||
		U.Name != "42" ||
		U.Projects[1] != "bar" ||
		U.Projects[0] != "foo" ||
		U.Admins[1] != "go" ||
		U.Members[1] != "2" ||
		U.Members[0] != "1" {
		t.Error("Invalid org, mismatched data", U)
	}

	m := U.ToMap()
	uu := NewOrganizationFromMap(m)
	if uu.Owners[0] != "mudler" || uu.Owners[1] != "borg" ||
		uu.Name != "42" ||
		uu.Projects[1] != "bar" ||
		uu.Projects[0] != "foo" ||
		uu.Members[1] != "2" ||
		uu.Members[0] != "1" {
		t.Error("Invalid org, mismatched data", U)
	}

	if !uu.ContainsOwner("mudler") {
		t.Error("Invalid owners", uu)
	}
	if uu.ContainsOwner("invalid") {
		t.Error("Invalid owners", uu)
	}
	if !uu.ContainsMember("2") {
		t.Error("Invalid members", uu)
	}
	if uu.ContainsMember("invalid") {
		t.Error("Invalid members", uu)
	}
	if !uu.ContainsAdmin("go") {
		t.Error("Invalid admin", uu)
	}
	if uu.ContainsAdmin("invalid") {
		t.Error("Invalid admin", uu)
	}

	uu.AddAdmin("test")
	uu.AddOwner("test2")
	uu.AddMember("test3")
	if !uu.ContainsAdmin("test") {
		t.Error("Invalid admin", uu)
	}
	if !uu.ContainsOwner("test2") {
		t.Error("Invalid owners", uu)
	}
	if !uu.ContainsMember("test3") {
		t.Error("Invalid members", uu)
	}
}
