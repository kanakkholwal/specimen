declare global {
	namespace App {
		interface Error {
			code?: string;
		}
		interface Locals {}
		interface PageData {}
		interface PageState {}
		interface Platform {
			env: {
				DB: D1Database;
				ASSETS_BUCKET: R2Bucket;
			};
			context: {
				waitUntil(promise: Promise<unknown>): void;
			};
			caches: CacheStorage & { default: Cache };
		}
	}
}

export {};
