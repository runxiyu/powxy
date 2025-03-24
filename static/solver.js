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

let wasm_instance = null;
let wasm_exports = null;

async function load_wasm() {
	let response = await fetch("/.powxy/wasm/solver.wasm");
	let { instance } = await WebAssembly.instantiateStreaming(response, {
		env: {
			memory: new WebAssembly.Memory({ initial: 1 })
		}
	});
	wasm_instance = instance;
	wasm_exports = instance.exports;
}

onmessage = async function(e) {
	let { identifier_bytes, difficulty } = e.data;

	if (!wasm_instance) {
		await load_wasm();
	}

	let ptr = wasm_exports.get_challenge_ptr();
	let memory = new Uint8Array(wasm_instance.exports.memory.buffer, ptr, 32);
	memory.set(identifier_bytes);

	let nonce = wasm_exports.solve(difficulty);

	let buf = new ArrayBuffer(8);
	let view = new DataView(buf);
	view.setBigUint64(0, BigInt(nonce), true);
	let nonce_bytes = new Uint8Array(buf);

	postMessage({ nonce, nonce_bytes });
};
