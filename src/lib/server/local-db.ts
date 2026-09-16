import { existsSync } from 'node:fs';
import { resolve } from 'node:path';

// A D1-shaped façade over the SQLite bundle that `supply export --sqlite` writes, so local
// dev reads the same content as production without wrangler or miniflare state.
const BUNDLE = resolve(process.cwd(), 'dist/specimen.db');

type Row = Record<string, unknown>;
type Runner = { all: (...params: unknown[]) => Row[] };

interface LocalStatement {
	bind(...values: unknown[]): LocalStatement;
	all<T>(): Promise<{ results: T[] }>;
	first<T>(): Promise<T | null>;
}

// Vite's SSR loader rewrites a normal dynamic import and then fails to resolve a runtime
// builtin, so this builds one the transform cannot see.
const nativeImport = new Function('s', 'return import(s)') as (s: string) => Promise<unknown>;

function statement(prepare: (sql: string) => Runner, sql: string): LocalStatement {
	let bound: unknown[] = [];
	const stmt: LocalStatement = {
		bind(...values: unknown[]) {
			bound = values;
			return stmt;
		},
		async all<T>() {
			return { results: prepare(sql).all(...bound) as T[] };
		},
		async first<T>() {
			const rows = prepare(sql).all(...bound) as T[];
			return rows.length ? rows[0] : null;
		}
	};
	return stmt;
}

async function openDatabase(): Promise<(sql: string) => Runner> {
	const isBun = typeof (globalThis as { Bun?: unknown }).Bun !== 'undefined';

	if (isBun) {
		const mod = (await nativeImport('bun:sqlite')) as {
			Database: new (p: string, o?: { readonly?: boolean }) => { query: (s: string) => Runner };
		};
		const db = new mod.Database(BUNDLE, { readonly: true });
		return (sql: string) => db.query(sql);
	}

	const mod = (await nativeImport('node:sqlite')) as {
		DatabaseSync: new (p: string, o?: { readOnly?: boolean }) => { prepare: (s: string) => Runner };
	};
	const db = new mod.DatabaseSync(BUNDLE, { readOnly: true });
	return (sql: string) => db.prepare(sql);
}

let cached: unknown = null;

export async function openLocalD1(): Promise<unknown> {
	if (cached) return cached;
	if (!existsSync(BUNDLE)) {
		throw new Error(
			`Local index missing at ${BUNDLE}. Build it with:\n` +
				`  cd server && go run ./cmd/supply export --sqlite ../dist/specimen.db`
		);
	}

	const prepare = await openDatabase();

	cached = {
		prepare: (sql: string) => statement(prepare, sql),
		batch: async (statements: LocalStatement[]) => Promise.all(statements.map((s) => s.all())),
		dump: async () => new ArrayBuffer(0),
		exec: async () => ({ count: 0, duration: 0 })
	};
	return cached;
}
