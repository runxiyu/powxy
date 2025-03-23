// SPDX-License-Identifier: AGPL-3.0-only
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
	tmpl, err = template.New("powxy").Parse(`
<!DOCTYPE html>
<html>
<head>
<title>Proof of Work Challenge</title>
<script>
/*    
@licstart  The following is the entire license notice for the 
JavaScript code in this page.

Copyright (C) 2025  Runxi Yu

The JavaScript code in this page is free software: you can
redistribute it and/or modify it under the terms of the GNU
General Public License (GNU GPL) as published by the Free Software
Foundation, either version 3 of the License, or (at your option)
any later version.  The code is distributed WITHOUT ANY WARRANTY;
without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE.  See the GNU GPL for more details.

As additional permission under GNU GPL version 3 section 7, you
may distribute non-source (e.g., minimized or compacted) forms of
that code without the copy of the GNU GPL normally required by
section 4, provided you include this license notice and a URL
through which recipients can access the Corresponding Source.   


@licend  The above is the entire license notice
for the JavaScript code in this page.
*/
</script>
</head>
<body>
<header>
<h1>Proof of Work Challenge</h1>
</header>
<p>This site is protected by Powxy{{ if .Global.Version }} {{ .Global.Version }}{{ end }}.</p>
<p>You must complete this proof of work challenge before you could access this site.</p>
{{- if .Message }}
<p><strong>{{ .Message }}</strong></p>
{{- end }}
<p>Select an nonce shorter than or equal to 32 bytes, such that when it is appended to the decoded form of the following base64 string, and a SHA-256 hash is taken as a whole, the first {{ .Global.NeedBits }} bits of the SHA-256 hash are zeros. Within one octet, higher bits are considered to be in front of lower bits.</p>
<p>{{ .UnsignedTokenBase64 }}</p>
<form method="POST">
<p>
Encode your selected nonce in base64 and submit it below.
</p>
<p>
Please note that if your submission is successful, you will be given a cookie that will allow you to access this site for a period of time without having to complete the challenge again. By pressing the submit button, you agree to be given cookies for this purpose.
</p>
<input name="powxy" type="text" />
<input type="submit" value="Submit" />
</form>
<p id="solver_status"></p>
<details>
<summary>Offline solver program</summary>
<pre>` + html.EscapeString(solverProgram) + `</pre>
</details>
<p>
Powxy is free software: you can redistribute it and/or modify it under the terms of the <a href="https://www.gnu.org/licenses/agpl-3.0.html">GNU Affero General Public License, version 3</a>, as published by the Free Software Foundation. Powxy is distributed in the hope that it will be useful, but <strong>without any warranty</strong>; without even the implied warranty of <strong>merchantability</strong> or <strong>fitness for a particular purpose</strong>. See the GNU Affero General Public License for more details. The source code is available at {{ .Global.SourceURL }}
</p>
</body>
<script>
document.addEventListener("DOMContentLoaded", function() {
	let challenge_b64 = "{{ .UnsignedTokenBase64 }}";
	let difficulty = {{ .Global.NeedBits }};
	let form = document.querySelector("form");
	let field = form.querySelector("input[name='powxy']");
	let status_el = document.getElementById("solver_status");

	if (!window.crypto || !window.crypto.subtle) {
		status_el.textContent = "SubtleCrypto not available. You must solve the challenge externally.";
		return;
	}
	status_el.textContent = "SubtleCrypto detected. Attempting to solve in JS...";

	let solver_active = true;
	form.addEventListener("submit", function() {
		solver_active = false;
	});

	async function solve_pow() {
		let token_bytes = Uint8Array.from(
			atob(challenge_b64),
			ch => ch.charCodeAt(0)
		);

		let nonce = 0n;
		let buf = new ArrayBuffer(8);
		let view = new DataView(buf);

		while (solver_active) {
			view.setBigUint64(0, nonce, true);

			let candidate = new Uint8Array(token_bytes.length + 8);
			candidate.set(token_bytes, 0);
			candidate.set(new Uint8Array(buf), token_bytes.length);

			let digest_buffer = await crypto.subtle.digest("SHA-256", candidate);
			let digest = new Uint8Array(digest_buffer);

			if (has_leading_zero_bits(digest, difficulty)) {
				let nonce_str = String.fromCharCode(...new Uint8Array(buf));
				field.value = btoa(nonce_str);

				status_el.textContent = "Solution found.";
				return;
			}

			nonce++;

			// Update status every 256 tries
			if ((nonce & 0x00FFn) === 0n) {
				status_el.textContent = "Attempting to solve automatically via SubtleCrypto. Tried " + nonce + " candidates so far...";
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
</html>
`)
	if err != nil {
		log.Fatal(err)
	}
}
