// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

<<<<<<< HEAD
//go:build goexperiment.unified
// +build goexperiment.unified
=======
//go:build go1.18 && goexperiment.unified
// +build go1.18,goexperiment.unified
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)

package gcimporter

const unifiedIR = true
