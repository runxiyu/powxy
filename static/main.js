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
		status_el.textContent = "Challenge solved automatically in " + nonce + " iterations";
	};

	worker.postMessage({ identifier_bytes, difficulty });
});
