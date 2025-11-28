import { AsyncLocalStorage } from "async_hooks";
import * as dotenv from "dotenv";
import pg from "pg";

dotenv.config();

export const pool = new pg.Pool({
	connectionString: process.env.DATABASE_URL,
	max: 8,
});

pool.on("error", (error) => console.error(error));

pool.on("connect", (client) => {
	client.on("error", (error) => console.error(error));
});

/**
 * A pool client whose lifecycle is managed.
 */
type ManagedPoolClient = Omit<pg.PoolClient, "release">;

/**
 * A pool client that automatically releases itself when disposed.
 */
type DisposablePoolClient = ManagedPoolClient & Disposable;

// polyfill if needed
if (!Object.hasOwn(Symbol, "dispose")) {
	Object.defineProperty(Symbol, "dispose", { value: Symbol("dispose") });
}

/**
 * Returns a pool client.
 * If called in {@link transaction} or {@link stickyClient}, it will return the client made for that context.
 * If not, it will return a fresh client that is disposed at the end of the call.
 *
 * Direct use is discouraged unless you need to run raw queries.
 */
export async function useClient(): Promise<DisposablePoolClient> {
	const storedClient = getStoredClient();

	if (storedClient) {
		// the stored client is already managed, so we should do nothing on its disposal.
		// to make sure that it can still be disposed in the end, back up the original dispose function and restore it when it is disposed locally.
		const dispose = storedClient.client[Symbol.dispose];

		storedClient.client[Symbol.dispose] = () => {
			storedClient.client[Symbol.dispose] = dispose;
		};

		return storedClient.client;
	}

	const client = await pool.connect();
	let released = false;
	return Object.assign(client, {
		[Symbol.dispose]: () => {
			if (released) return;
			client.release();
			released = true;
		},
	});
}

type StoredClient = {
	transaction: {
		id: string;
	} | null;

	/**
	 * Acquired database client.
	 * Set to null if disposed.
	 */
	client: DisposablePoolClient | null;
};

const asyncLocalStorage = new AsyncLocalStorage<StoredClient | undefined>();

export function runWithStoredClient<T>(
	storedClient: StoredClient & { client: DisposablePoolClient },
	fn: () => T,
) {
	const _storedClient: StoredClient = storedClient;

	return asyncLocalStorage.run(_storedClient, () =>
		Promise.resolve()
			.then(fn)
			.finally(() => {
				_storedClient.client = null;
			}),
	);
}

function getStoredClient():
	| (StoredClient & { client: DisposablePoolClient })
	| undefined {
	const storedClient = asyncLocalStorage.getStore();

	if (!storedClient) {
		return undefined;
	}

	const { client } = storedClient;

	if (client === null) {
		// We are in a context where a client was used, but it has already been released.
		// This is likely a result of creating background jobs within transactions or stickyClients.
		throw new Error("Cannot use a released client.");
	}

	return { ...storedClient, client };
}

export function getTransaction() {
	return getStoredClient()?.transaction ?? null;
}

export function inTransaction(): boolean {
	return getTransaction() !== null;
}

export function getTransactionId(): string | null {
	return getTransaction()?.id ?? null;
}

/**
 * Runs all queries with the same pool client.
 * Similar to {@link transaction}, but does not start a transaction.
 * Not really needed in most cases, but can improve performance when running a lot of queries.
 */
export async function stickyClient<T>(fn: () => Promise<T>): Promise<T> {
	if (inTransaction()) {
		// this is technically possible, but would make it possible to have nested transactions due to transactional becoming false
		// nevertheless, using a sticky client inside a transaction is probably a mistake
		throw new Error("Cannot use stickyClient inside a transaction");
	}

	using client = await useClient();
	return await runWithStoredClient({ transaction: null, client }, fn);
}

/**
 * Gets rid of a transactional context.
 * Useful for things that should never be rolled back, e.g. incrementing a counter.
 */
export async function escapeTransaction<T>(fn: () => Promise<T>): Promise<T> {
	return await asyncLocalStorage.exit(fn);
}
