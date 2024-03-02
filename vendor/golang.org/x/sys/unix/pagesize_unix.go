// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

<<<<<<< HEAD
//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
=======
//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

// For Unix, get the pagesize from the runtime.

package unix

import "syscall"

func Getpagesize() int {
	return syscall.Getpagesize()
}
