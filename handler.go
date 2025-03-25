// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// handler handles an incoming HTTP request.
func handler(writer http.ResponseWriter, request *http.Request) {
	remoteIP := getRemoteIP(request)
	userAgent := request.Header.Get("User-Agent")
	uri := request.RequestURI

	// Static resources for powxy itself.
	if strings.HasPrefix(request.URL.Path, "/.powxy/") {
		http.StripPrefix("/.powxy/", http.FileServer(http.FS(resourcesFS))).ServeHTTP(writer, request)
		return
	}

	// We attempt to fetch the powxy cookie. Its non-existence
	// does not matter here; if the cookie does not exist, it
	// will be nil, so validation will simply fail and the user
	// will be prompted to solve the PoW challenge.
	cookie, err := request.Cookie("powxy")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		slog.Error("\x0304ERRCOOKIE",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"error", err,
		)
		http.Error(writer, "error fetching cookie", http.StatusInternalServerError)
		return
	}

	// We generate the identifier that identifies the client,
	// and the expected HMAC that the cookie should include.
	identifier, expectedMAC := makeIdentifierMAC(request)

	// If the cookie exists and is valid, we simply proxy the
	// request.
	if validateCookie(cookie, expectedMAC) {
		slog.Info("\x0302PROXY",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
		)
		proxyRequest(writer, request)
		return
	}

	// A convenience function to render the challenge page,
	// since all parameters but the message are constant at this
	// point.
	challengePage := func(message string) {
		writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0, private")
		writer.Header().Set("Pragma", "no-cache")
		writer.Header().Set("Expires", "0")
		err := tmpl.Execute(writer, tparams{
			Identifier: base64.StdEncoding.EncodeToString(identifier),
			Message:    message,
			Global:     global,
		})
		if err != nil {
			slog.Error("\x0304template execution failed",
				"ip", remoteIP,
				"uri", uri,
				"user_agent", userAgent,
				"error", err,
			)
		}
	}

	// This generally shouldn't happen, at least not for web
	// browesrs.
	err = request.ParseForm()
	if err != nil {
		slog.Warn("\x0304MALFORMED",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"error", err,
		)
		challengePage("You submitted a malformed form.")
		return
	}

	formValues, ok := request.PostForm["powxy"]
	if !ok {
		// If there's simply no form value, the user is probably
		// just visiting the site for the first time or with an
		// expired cookie.
		slog.Info("\x0301POW CHL",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
		)
		challengePage("")
		return
	} else if len(formValues) != 1 {
		// This should never happen, at least not for web
		// browsers.
		slog.Warn("\x0304FORMNUM",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"form_values", formValues,
		)
		challengePage("You submitted an invalid number of form values.")
		return
	}

	// We validate that the length is reasonable before even
	// decoding it with base64.
	if len(formValues[0]) > 43 {
		slog.Warn("\x0304TOOLONG",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"form_value", formValues[0],
		)
		challengePage("Your submission was too long.")
		return
	}

	// Actually decode the base64 value.
	nonce, err := base64.StdEncoding.DecodeString(formValues[0])
	if err != nil {
		slog.Warn("\x0304ERRBASE64",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"error", err,
			"form_value", formValues[0],
		)
		challengePage("Your submission was improperly encoded.")
		return
	}

	// Validate the nonce.
	if !validateNonce(identifier, nonce) {
		slog.Warn("\x0304",
			"ip", remoteIP,
			"uri", uri,
			"user_agent", userAgent,
			"form_value", formValues[0],
		)
		challengePage("Your submission was incorrect, or your session has expired while submitting.")
		return
	}

	// Everything starting here: the nonce is valid, and we
	// can set the cookie and redirect them. The redirection is
	// needed as their "normal" request is most definitely
	// different from one to expect after solving the PoW
	// challenge.

	http.SetCookie(writer, &http.Cookie{
		Name:     "powxy",
		Value:    base64.StdEncoding.EncodeToString(expectedMAC),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	slog.Info("\x0303POW ACK",
		"ip", remoteIP,
		"uri", uri,
		"user_agent", userAgent,
		"form_value", formValues[0],
	)
	http.Redirect(writer, request, "", http.StatusSeeOther)
}

// tparams holds parameters for the template.
type tparams struct {
	Identifier string
	Message    string
	Global     any
}
