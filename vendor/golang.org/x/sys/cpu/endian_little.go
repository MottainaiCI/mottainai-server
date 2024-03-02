// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64 || amd64p32 || alpha || arm || arm64 || loong64 || mipsle || mips64le || mips64p32le || nios2 || ppc64le || riscv || riscv64 || sh || wasm
<<<<<<< HEAD
<<<<<<< HEAD
=======
// +build 386 amd64 amd64p32 alpha arm arm64 loong64 mipsle mips64le mips64p32le nios2 ppc64le riscv riscv64 sh wasm
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

package cpu

// IsBigEndian records whether the GOARCH's byte order is big endian.
const IsBigEndian = false
