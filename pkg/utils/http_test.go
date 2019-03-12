/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>
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

package utils

import "testing"

func TestPathEscape(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Simple path", args{path: "/test/"}, "/test/"},
		{"Simple path", args{path: "/test"}, "/test"},
		{"Simple path", args{path: "test/"}, "test/"},

		{"Complex path", args{path: "/test:try/#then/:false!"}, "/test%3Atry/%23then/%3Afalse%21"},
		{"Complex path", args{path: "test:try/#then/:false!"}, "test%3Atry/%23then/%3Afalse%21"},
		{"Complex path", args{path: "test:try/#then/:false!/"}, "test%3Atry/%23then/%3Afalse%21/"},
		{"Complex path", args{path: "/test:try/#then/:false!/"}, "/test%3Atry/%23then/%3Afalse%21/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathEscape(tt.args.path); got != tt.want {
				t.Errorf("PathEscape(%v) = %v, want %v", tt.args.path, got, tt.want)
			}
		})
	}
}
