// SPDX-License-Identifier: BSD-2-Clause
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"html"
	"log"
	"text/template"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("powxy").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>Proof-of-work challenge</title>
	<style>
		html {
			font-family: sans-serif;
			background-color: var(--background-color);
			color: var(--text-color);
			--radius-1: 0.32rem;
			--background-color: hsl(0, 0%, 100%);
			--text-color: hsl(0, 0%, 0%);
			--link-color: hsl(320, 50%, 36%);
			--light-text-color: hsl(0, 0%, 45%);
			--darker-border-color: hsl(0, 0%, 72%);
			--lighter-border-color: hsl(0, 0%, 85%);
			--text-decoration-color: hsl(0, 0%, 72%);
			--darker-box-background-color: hsl(0, 0%, 92%);
			--lighter-box-background-color: hsl(0, 0%, 95%);
			--primary-color: hsl(320, 50%, 36%);
			--primary-color-contrast: hsl(320, 0%, 100%);
			--danger-color: #ff0000;
			--danger-color-contrast: #ffffff;
		}

		@media (prefers-color-scheme: dark) {
			html {
				--background-color: hsl(0, 0%, 0%);
				--text-color: hsl(0, 0%, 100%);
				--link-color: hsl(320, 50%, 76%);
				--light-text-color: hsl(0, 0%, 78%);
				--darker-border-color: hsl(0, 0%, 35%);
				--lighter-border-color: hsl(0, 0%, 25%);
				--text-decoration-color: hsl(0, 0%, 30%);
				--darker-box-background-color: hsl(0, 0%, 20%);
				--lighter-box-background-color: hsl(0, 0%, 15%);
			}
		}

		body {
			margin: 0;
			padding: 1rem;
		}

		main {
			max-width: 720px;
			margin: 0 auto;
		}

		*:focus-visible {
			outline: 1.5px var(--primary-color) solid;
		}

		section {
			margin: 0;
		}

		label {
			display: block;
			font-style: italic;
			margin-top: 1rem;
			margin-bottom: 0.5rem;
		}

		h1 {
			margin-top: 0;
			color: var(--primary-color);
		}

		p, summary {
			line-height: 1.2;
			font-size: 1rem;
		}

		a {
			color: var(--link-color);
			text-decoration-color: var(--text-decoration-color);
		}

		input[type="text"] {
			font-family: monospace;
			font-size: 1rem;
			background-color: var(--lighter-box-background-color);
			width: 100%;
			padding: 0.5rem;
			border-radius: var(--radius-1);
			border: none;
			box-shadow: inset 0 1px 4px rgba(0, 0, 0, 0.15);
			margin-bottom: 1rem;
			box-sizing: border-box;
		}

		input[type="submit"] {
			padding: 0.5rem 1rem;
			background-color: var(--primary-color);
			color: var(--primary-color-contrast);
			border: none;
			border-radius: var(--radius-1);
			cursor: pointer;
		}

		input[readonly] {
			background-color: var(--lighter-box-background-color);
			color: var(--text-color);
			cursor: text;
		}

		details {
			margin-top: 2rem;
			background-color: var(--lighter-box-background-color);
			padding: 0.5rem;
			border-radius: var(--radius-1);
			box-shadow: inset 0 1px 4px rgba(0, 0, 0, 0.15);
		}

		pre {
			overflow-x: auto;
			white-space: pre-wrap;
			word-break: break-word;
		}

		#solver_status {
			color: var(--light-text-color);
			margin-top: 1rem;
		}
	</style>
	<script>
		/*    
		@licstart  The following is the entire license notice for the 
		JavaScript code in this page.
		
		Copyright (C) 2025  Runxi Yu
		
		Redistribution and use in source and binary forms, with or without
		modification, are permitted provided that the following conditions are
		met:
		
		    1. Redistributions of source code must retain the above copyright
		    notice, this list of conditions and the following disclaimer.
		
		    2. Redistributions in binary form must reproduce the above copyright
		    notice, this list of conditions and the following disclaimer in the
		    documentation and/or other materials provided with the distribution.
		
		THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
		EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
		IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
		PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
		CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
		EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
		PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
		PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
		LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
		NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
		SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

		@licend  The above is the entire license notice
		for the JavaScript code in this page.
		*/
	</script>
</head>
<body>
	<main>
		<header>
			<h1>Proof-of-work challenge</h1>
		</header>

		<section>
			<p>This site is protected by <a href="{{ .Global.SourceURL }}">Powxy</a>{{ if .Global.Version }} {{ .Global.Version }}{{ end }}.</p>
			<p>You must complete this proof-of-work challenge before you can access this site.</p>
		</section>

		{{- if .Message }}
		<section>
			<p><strong>{{ .Message }}</strong></p>
		</section>
		{{- end }}

		<section>
			<p>Select a nonce no longer than 32 bytes, such that when it is appended to the decoded form of the challenge identifier, and the entire result is hashed with SHA-256, the first {{ .Global.NeedBits }} bits of the SHA-256 hash are all zeros. Within one octet, higher bits are considered to come before lower bits.</p>
			<label for="unsigned-identifier">Challenge identifier (read-only)</label>
			<input id="unsigned-identifier" type="text" readonly disabled tabindex="-1" value="{{ .Identifier }}" />
		</section>

		<section>
			<form method="POST">
				<p>Encode your selected nonce in base64 and submit it below.</p>
				<p>Please note that if your submission is successful, you will be given a cookie that will allow you to access this site for a period of time without having to complete the challenge again. By pressing the submit button, you agree to be given cookies for this purpose.</p>
				<label for="nonce">Nonce</label>
				<input id="nonce" name="powxy" type="text" />
				<input type="submit" value="Submit" />
			</form>
		</section>

		<section>
			<p id="solver_status">JavaScript seems to be disabled. You must solve the challenge externally.</p>
		</section>

		<details>
			<summary>Offline solver program</summary>
			<pre>` + html.EscapeString(solverProgram) + `</pre>
		</details>
	</main>

	<script>
		document.addEventListener("DOMContentLoaded", function() {
			let challenge_b64 = "{{ .Identifier }}";
			let difficulty = {{ .Global.NeedBits }};
			let form = document.querySelector("form");
			let field = form.querySelector("input[name='powxy']");
			let status_el = document.getElementById("solver_status");
	
			if (!window.crypto || !window.crypto.subtle) {
				status_el.textContent = "SubtleCrypto not available. You must solve the challenge externally.";
				return;
			}
			status_el.textContent = "SubtleCrypto detected. Attempting to solve the challenge automatically...";
	
			let solver_active = true;
			form.addEventListener("submit", function() {
				solver_active = false;
			});
	
			async function solve_pow() {
				let identifier_bytes = Uint8Array.from(
					atob(challenge_b64),
					ch => ch.charCodeAt(0)
				);
	
				let nonce = 0n;
				let buf = new ArrayBuffer(8);
				let view = new DataView(buf);
	
				while (solver_active) {
					view.setBigUint64(0, nonce, true);
	
					let candidate = new Uint8Array(identifier_bytes.length + 8);
					candidate.set(identifier_bytes, 0);
					candidate.set(new Uint8Array(buf), identifier_bytes.length);
	
					let digest_buffer = await crypto.subtle.digest("SHA-256", candidate);
					let digest = new Uint8Array(digest_buffer);
	
					if (has_leading_zero_bits(digest, difficulty)) {
						let nonce_str = String.fromCharCode(...new Uint8Array(buf));
						field.value = btoa(nonce_str);
	
						status_el.textContent = "A solution has been found automatically in " + nonce + " iterations.";
						return;
					}
	
					nonce++;
	
					if ((nonce & 0x00FFn) === 0n) {
						status_el.textContent = "Attempting to solve automatically. Tried " + nonce + " candidates so far...";
						await new Promise(r => setTimeout(r, 0));
					}
				}
			}
	
			function has_leading_zero_bits(digest, bits) {
				let full_bytes = bits >>> 3;
				for (let i = 0; i < full_bytes; i++) {
					if (digest[i] !== 0) {
						return false;
					}
				}
				let remainder = bits & 7;
				if (remainder !== 0) {
					let mask = 0xFF << (8 - remainder);
					if ((digest[full_bytes] & mask) !== 0) {
						return false;
					}
				}
				return true;
			}
	
			solve_pow();
		});
	</script>
</body>
</html>`)
	if err != nil {
		log.Fatal(err)
	}
}
