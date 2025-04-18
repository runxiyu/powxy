// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	_ "embed"
	"log"
	"text/template"
)

//go:embed challenge.html
var tmplString string
var tmpl *template.Template

// This init function parses the HTML template.
func init() {
	var err error
	tmpl, err = template.New("powxy").Parse(tmplString)
	if err != nil {
		log.Fatal(err)
	}
}
