// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package fsutils provides file system admin commands.
package fsutils

import (
	"github.com/platinasystems/go/fsutils/mount"
	"github.com/platinasystems/go/fsutils/umount"
)

func New() []interface{} {
	return []interface{}{
		mount.New(),
		umount.New(),
	}
}