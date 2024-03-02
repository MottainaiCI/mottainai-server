// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package packagesdriver fetches type sizes for go/packages and go/analysis.
package packagesdriver

import (
	"context"
	"fmt"
<<<<<<< HEAD
<<<<<<< HEAD
=======
	"go/types"
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	"strings"

	"golang.org/x/tools/internal/gocommand"
)

<<<<<<< HEAD
<<<<<<< HEAD
func GetSizesForArgsGolist(ctx context.Context, inv gocommand.Invocation, gocmdRunner *gocommand.Runner) (string, string, error) {
=======
var debug = false

func GetSizesGolist(ctx context.Context, inv gocommand.Invocation, gocmdRunner *gocommand.Runner) (types.Sizes, error) {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
func GetSizesForArgsGolist(ctx context.Context, inv gocommand.Invocation, gocmdRunner *gocommand.Runner) (string, string, error) {
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	inv.Verb = "list"
	inv.Args = []string{"-f", "{{context.GOARCH}} {{context.Compiler}}", "--", "unsafe"}
	stdout, stderr, friendlyErr, rawErr := gocmdRunner.RunRaw(ctx, inv)
	var goarch, compiler string
	if rawErr != nil {
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		rawErrMsg := rawErr.Error()
		if strings.Contains(rawErrMsg, "cannot find main module") ||
			strings.Contains(rawErrMsg, "go.mod file not found") {
			// User's running outside of a module.
			// All bets are off. Get GOARCH and guess compiler is gc.
<<<<<<< HEAD
=======
		if rawErrMsg := rawErr.Error(); strings.Contains(rawErrMsg, "cannot find main module") || strings.Contains(rawErrMsg, "go.mod file not found") {
			// User's running outside of a module. All bets are off. Get GOARCH and guess compiler is gc.
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			// TODO(matloob): Is this a problem in practice?
			inv.Verb = "env"
			inv.Args = []string{"GOARCH"}
			envout, enverr := gocmdRunner.Run(ctx, inv)
			if enverr != nil {
<<<<<<< HEAD
<<<<<<< HEAD
				return "", "", enverr
			}
			goarch = strings.TrimSpace(envout.String())
			compiler = "gc"
		} else if friendlyErr != nil {
			return "", "", friendlyErr
		} else {
			// This should be unreachable, but be defensive
			// in case RunRaw's error results are inconsistent.
			return "", "", rawErr
=======
				return nil, enverr
=======
				return "", "", enverr
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			}
			goarch = strings.TrimSpace(envout.String())
			compiler = "gc"
		} else if friendlyErr != nil {
			return "", "", friendlyErr
		} else {
<<<<<<< HEAD
			return nil, friendlyErr
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
			// This should be unreachable, but be defensive
			// in case RunRaw's error results are inconsistent.
			return "", "", rawErr
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		}
	} else {
		fields := strings.Fields(stdout.String())
		if len(fields) < 2 {
<<<<<<< HEAD
<<<<<<< HEAD
			return "", "", fmt.Errorf("could not parse GOARCH and Go compiler in format \"<GOARCH> <compiler>\":\nstdout: <<%s>>\nstderr: <<%s>>",
=======
			return nil, fmt.Errorf("could not parse GOARCH and Go compiler in format \"<GOARCH> <compiler>\":\nstdout: <<%s>>\nstderr: <<%s>>",
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
			return "", "", fmt.Errorf("could not parse GOARCH and Go compiler in format \"<GOARCH> <compiler>\":\nstdout: <<%s>>\nstderr: <<%s>>",
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
				stdout.String(), stderr.String())
		}
		goarch = fields[0]
		compiler = fields[1]
	}
<<<<<<< HEAD
<<<<<<< HEAD
	return compiler, goarch, nil
=======
	return types.SizesFor(compiler, goarch), nil
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	return compiler, goarch, nil
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}
