// Copyright 2020 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package starlarkjson is an alias for github.com/spachava753/starlarkx/lib/json to provide
// backwards compatibility
//
// Deprecated: use github.com/spachava753/starlarkx/lib/json instead
package starlarkjson // import "github.com/spachava753/starlarkx/starlarkjson"

import (
	"github.com/spachava753/starlarkx/lib/json"
)

// Module is an alias of json.Module for backwards import compatibility
var Module = json.Module
