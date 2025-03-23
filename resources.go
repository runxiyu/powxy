// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"embed"
)

//go:embed wasm/*.wasm
//go:embed static/*
var resourcesFS embed.FS
