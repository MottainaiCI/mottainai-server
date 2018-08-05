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

package user

import "testing"

func TestNewUserFromMap(t *testing.T) {

	U := NewUserFromMap(map[string]interface{}{"name": "waiting", "password": "none"})

	if U.Name != "waiting" {
		t.Error("Invalid username", U.Name)
	}
	if U.Password != "none" {
		t.Error("Invalid password", U.Password)
	}

}

func TestPasswordSalt(t *testing.T) {
	U2 := NewUserFromMap(map[string]interface{}{"name": "test", "password": "test2"})
	if U2.Password != "test2" {
		t.Error("Password was not read", U2.Password)
	}

	err := U2.SaltPassword()
	if U2.Password == "test2" || err != nil {
		t.Error("Password was not salted", U2.Password, err)
	}

	ok, newHash := U2.VerifyPassword("test2")
	if !ok {
		t.Error("Password Not verified", ok, newHash)
	}
}

func TestIdentities(t *testing.T) {
	u := NewUserFromMap(map[string]interface{}{"name": "test", "password": "test2"})
	u.AddIdentity("github", &Identity{ID: "foo"})
	if u.Identities["github"].ID != "foo" {
		t.Fatal("Identity not added", u.Identities)
	}
	u.RemoveIdentity("github")

	if _, ok := u.Identities["github"]; ok {
		t.Fatal("Identity still present")
	}
	u.RemoveIdentity("github")
}
