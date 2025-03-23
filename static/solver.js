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
