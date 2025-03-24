/*    
 * @licstart  The following is the entire license notice for the 
 * JavaScript code in this page.
 * 
 * Copyright (C) 2025  Runxi Yu
 * 
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 * 
 *     1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 * 
 *     2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 * 
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * 
 * @licend  The above is the entire license notice
 * for the JavaScript code in this page.
 */

document.addEventListener("DOMContentLoaded", async function() {
	let challenge_b64 = document.body.dataset.identifier;
	let difficulty = parseInt(document.body.dataset.difficulty);
	let form = document.querySelector("form");
	let field = form.querySelector("input[name='powxy']");
	let status_el = document.getElementById("solver_status");

	let identifier_bytes = Uint8Array.from(
		atob(challenge_b64),
		ch => ch.charCodeAt(0)
	);

	status_el.textContent = "Starting WebAssembly solver as a worker...";

	let worker = new Worker("/.powxy/static/solver.js");

	worker.onmessage = function(e) {
		let { nonce, nonce_bytes } = e.data;
		let nonce_str = String.fromCharCode(...nonce_bytes);
		field.value = btoa(nonce_str);
		status_el.textContent = "Challenge solved automatically in " + nonce + " iterations.";
	};

	worker.postMessage({ identifier_bytes, difficulty });
});
