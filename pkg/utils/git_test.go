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

package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

var testurl = "https://github.com/MottainaiCI/mottainai-server"

func TestGitClone(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	repo, err := GitClone(testurl, tempdir)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Fatal("Repo should not be nil")
	}
	if _, err := os.Stat(tempdir + "/Makefile"); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestGitCheckout(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	repo, err := GitClone(testurl, tempdir)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Fatal("Repo should not be nil")
	}

	if _, err := os.Stat(tempdir + "/public/assets/js/lib/data-table/buttons.bootstrap.min.js"); !os.IsNotExist(err) {
		t.Fatal("Deleted file exists")
	}

	err = GitCheckoutCommit(repo, "9bed5fa4c05dc14e8e3f26539f847287c6d9248a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tempdir + "/public/assets/js/lib/data-table/buttons.bootstrap.min.js"); os.IsNotExist(err) {
		t.Fatal("File should exist now")
	}
}

func TestGitFetch(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "test2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	//os.RemoveAll(tempdir)

	repo, err := GitClone("https://github.com/Sabayon/for-gentoo", tempdir)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Fatal("Repo should not be nil")
	}
	// XXX: Doesn't work yet
	// err = GitFetch(repo, "origin", []string{"pull/75/head:CI_test"})
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = GitCheckoutCommit(repo, "CI_test")
	// if err != nil {
	// 	t.Fatal(err)
	// }

}
