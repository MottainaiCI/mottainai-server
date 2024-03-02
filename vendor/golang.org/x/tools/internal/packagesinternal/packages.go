// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package packagesinternal exposes internal-only fields from go/packages.
package packagesinternal

<<<<<<< HEAD
<<<<<<< HEAD
=======
import (
	"golang.org/x/tools/internal/gocommand"
)

>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
var GetForTest = func(p interface{}) string { return "" }
var GetDepsErrors = func(p interface{}) []*PackageError { return nil }

type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

<<<<<<< HEAD
<<<<<<< HEAD
=======
var GetGoCmdRunner = func(config interface{}) *gocommand.Runner { return nil }

var SetGoCmdRunner = func(config interface{}, runner *gocommand.Runner) {}

>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
var TypecheckCgo int
var DepsErrors int // must be set as a LoadMode to call GetDepsErrors
var ForTest int    // must be set as a LoadMode to call GetForTest

var SetModFlag = func(config interface{}, value string) {}
var SetModFile = func(config interface{}, value string) {}
