// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	_ "embed"
	"log"
	"text/template"
)

//go:embed challenge.tmpl
var tmplString string
var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("powxy").Parse(tmplString)
	if err != nil {
		log.Fatal(err)
	}
}
