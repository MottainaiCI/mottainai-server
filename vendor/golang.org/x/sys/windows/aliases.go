// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

<<<<<<< HEAD
//go:build windows
=======
//go:build windows && go1.9
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

package windows

import "syscall"

type Errno = syscall.Errno
type SysProcAttr = syscall.SysProcAttr
