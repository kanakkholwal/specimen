import fs from 'node:fs';
import path from 'node:path';

const ROOTS = ['src', 'scripts', '.notes'];
const EXTS = new Set(['.ts', '.js', '.mjs', '.svelte', '.css', '.html', '.md', '.json']);
// Built from char codes so this file does not trip its own gate.
const DASHES = new RegExp(`[${String.fromCharCode(0x2014, 0x2013)}]`);

// Vendored output, not code this project writes.
const GENERATED = /worker-configuration\.d\.ts$/;

function walk(dir, out = []) {
	if (!fs.existsSync(dir)) return out;
	for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
		const p = path.join(dir, entry.name);
		if (entry.isDirectory()) {
			if (entry.name === 'node_modules' || entry.name === '.svelte-kit') continue;
			walk(p, out);
		} else if (EXTS.has(path.extname(entry.name))) {
			out.push(p);
		}
	}
	return out;
}

const files = ROOTS.flatMap((r) => walk(r)).filter((f) => !GENERATED.test(f));
const hits = [];

for (const file of files) {
	const lines = fs.readFileSync(file, 'utf8').split(/\r?\n/);
	lines.forEach((line, i) => {
		if (DASHES.test(line)) hits.push(`${file}:${i + 1}: ${line.trim().slice(0, 100)}`);
	});
}

if (hits.length) {
	for (const h of hits) console.log(h);
	console.log(`\n${hits.length} em/en dash(es) in ${files.length} files`);
	process.exit(1);
}
console.log(`dash scan: clean (${files.length} files)`);
