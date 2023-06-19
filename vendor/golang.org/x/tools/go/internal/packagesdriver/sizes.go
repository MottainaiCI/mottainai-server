// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package packagesdriver fetches type sizes for go/packages and go/analysis.
package packagesdriver

import (
	"context"
	"fmt"
<<<<<<< HEAD
=======
	"go/types"
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
	"strings"

	"golang.org/x/tools/internal/gocommand"
)

<<<<<<< HEAD
func GetSizesForArgsGolist(ctx context.Context, inv gocommand.Invocation, gocmdRunner *gocommand.Runner) (string, string, error) {
=======
var debug = false

func GetSizesGolist(ctx context.Context, inv gocommand.Invocation, gocmdRunner *gocommand.Runner) (types.Sizes, error) {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
	inv.Verb = "list"
	inv.Args = []string{"-f", "{{context.GOARCH}} {{context.Compiler}}", "--", "unsafe"}
	stdout, stderr, friendlyErr, rawErr := gocmdRunner.RunRaw(ctx, inv)
	var goarch, compiler string
	if rawErr != nil {
<<<<<<< HEAD
		rawErrMsg := rawErr.Error()
		if strings.Contains(rawErrMsg, "cannot find main module") ||
			strings.Contains(rawErrMsg, "go.mod file not found") {
			// User's running outside of a module.
			// All bets are off. Get GOARCH and guess compiler is gc.
=======
		if rawErrMsg := rawErr.Error(); strings.Contains(rawErrMsg, "cannot find main module") || strings.Contains(rawErrMsg, "go.mod file not found") {
			// User's running outside of a module. All bets are off. Get GOARCH and guess compiler is gc.
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
			// TODO(matloob): Is this a problem in practice?
			inv.Verb = "env"
			inv.Args = []string{"GOARCH"}
			envout, enverr := gocmdRunner.Run(ctx, inv)
			if enverr != nil {
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
			}
			goarch = strings.TrimSpace(envout.String())
			compiler = "gc"
		} else {
			return nil, friendlyErr
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		}
	} else {
		fields := strings.Fields(stdout.String())
		if len(fields) < 2 {
<<<<<<< HEAD
			return "", "", fmt.Errorf("could not parse GOARCH and Go compiler in format \"<GOARCH> <compiler>\":\nstdout: <<%s>>\nstderr: <<%s>>",
=======
			return nil, fmt.Errorf("could not parse GOARCH and Go compiler in format \"<GOARCH> <compiler>\":\nstdout: <<%s>>\nstderr: <<%s>>",
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
				stdout.String(), stderr.String())
		}
		goarch = fields[0]
		compiler = fields[1]
	}
<<<<<<< HEAD
	return compiler, goarch, nil
=======
	return types.SizesFor(compiler, goarch), nil
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
}
