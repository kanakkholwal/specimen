import { spawnSync } from 'node:child_process';
import { existsSync, rmSync } from 'node:fs';

// svelte-check mirrors src into this folder and never prunes it, so a deleted route keeps
// failing the typecheck until it goes.
rmSync('.svelte-kit/.svelte-check', { recursive: true, force: true });

// This is the single source of truth for CI. The workflow calls this same script, so a
// green gate locally cannot turn red on push.
const steps = [
	{ name: 'comments', cmd: 'bun', args: ['scripts/check-comments.mjs', '--all'] },
	{ name: 'dashes', cmd: 'bun', args: ['scripts/check-dashes.mjs'] },
	{ name: 'lint', cmd: 'bunx', args: ['--bun', 'biome', 'lint', './src'] },
	{ name: 'format', cmd: 'bunx', args: ['--bun', 'biome', 'format', './src'] },
	// worker-configuration.d.ts is generated and untracked, so CI has to make it before typechecking.
	{ name: 'worker types', cmd: 'bun', args: ['run', 'types'] },
	{ name: 'types', cmd: 'bun', args: ['run', 'check'] },
	{ name: 'build', cmd: 'bun', args: ['run', 'build'], env: { SVELTE_KIT_OUT_DIR: '.svelte-kit/build' } }
];

if (existsSync('server/go.mod')) {
	steps.push(
		{ name: 'go vet', cmd: 'go', args: ['vet', './...'], cwd: 'server' },
		{ name: 'go test', cmd: 'go', args: ['test', './...'], cwd: 'server' },
		{ name: 'go comments', cmd: 'node', args: ['scripts/check-comments.mjs', '--all'], cwd: 'server' }
	);
}

const only = process.argv[2];
const selected = only ? steps.filter((s) => s.name.startsWith(only)) : steps;

let failed = 0;
const started = Date.now();

for (const step of selected) {
	const t = Date.now();
	process.stdout.write(`  ${step.name} ... `);
	const run = spawnSync(step.cmd, step.args, {
		cwd: step.cwd,
		env: { ...process.env, ...step.env },
		shell: true,
		encoding: 'utf8'
	});
	const ms = Date.now() - t;

	if (run.status === 0) {
		console.log(`ok (${(ms / 1000).toFixed(1)}s)`);
		continue;
	}

	failed++;
	console.log(`FAILED (${(ms / 1000).toFixed(1)}s)`);
	const output = `${run.stdout ?? ''}${run.stderr ?? ''}`.trim();
	for (const line of output.split(/\r?\n/).slice(-25)) console.log(`      ${line}`);
	console.log('');
}

const total = ((Date.now() - started) / 1000).toFixed(1);
if (failed) {
	console.log(`\ngate: ${failed} of ${selected.length} steps failed in ${total}s`);
	process.exit(1);
}
console.log(`\ngate: all ${selected.length} steps passed in ${total}s`);
