// import pg from 'pg';
// import { AsyncLocalStorage } from 'async_hooks';
//
// import dotenv from 'dotenv';
//
// dotenv.config();
//
// const pool = new pg.Pool({
//   connectionString: process.env.DATABASE_URL,
//   password: process.env.DATABASE_PASSWORD,
// });
// /**
//  * A pool client whose lifecycle is managed.
//  */
// type ManagedPoolClient = Omit<pg.PoolClient, 'release'>;
//
// /**
//  * A pool client that automatically releases itself when disposed.
//  */
// type DisposablePoolClient = ManagedPoolClient & Disposable;
//
// // polyfill if needed
// if (!Symbol.hasOwnProperty('dispose')) {
//   Object.defineProperty(Symbol, 'dispose', { value: Symbol('dispose') });
// }
//
// /**
//  * Returns a pool client.
//  * If called in {@link transaction} or {@link stickyClient}, it will return the client made for that context.
//  * If not, it will return a fresh client that is disposed at the end of the call.
//  *
//  * Direct use is discouraged unless you need to run raw queries.
//  */
// export async function useClient(): Promise<DisposablePoolClient> {
//   const storedClient = getStoredClient();
//
//   if (storedClient) {
//     // the stored client is already managed, so we should do nothing on its disposal.
//     // to make sure that it can still be disposed in the end, back up the original dispose function and restore it when it is disposed locally.
//     const dispose = storedClient.client[Symbol.dispose];
//
//     storedClient.client[Symbol.dispose] = () => {
//       storedClient.client[Symbol.dispose] = dispose;
//     };
//
//     return storedClient.client;
//   }
//
//   const client = await pool.connect();
//   let released = false;
//   return Object.assign(client, {
//     [Symbol.dispose]: () => {
//       if (released) return;
//       client.release();
//       released = true;
//     },
//   });
// }
//
// type StoredClient = {
//   transactional: boolean;
//   client: DisposablePoolClient;
// };
//
// const asyncLocalStorage = new AsyncLocalStorage<StoredClient>();
//
// export function runWithStoredClient<T>(
//   storedClient: StoredClient,
//   fn: () => T,
// ): T {
//   return asyncLocalStorage.run(storedClient, fn);
// }
//
// function getStoredClient(): StoredClient | undefined {
//   return asyncLocalStorage.getStore();
// }
//
// export function inTransaction(): boolean {
//   return getStoredClient()?.transactional === true;
// }
//
// /**
//  * Runs all queries with the same pool client.
//  * Similar to {@link transaction}, but does not start a transaction.
//  * Not really needed in most cases, but can improve performance when running a lot of queries.
//  */
// export async function stickyClient<T>(fn: () => Promise<T>): Promise<T> {
//   if (inTransaction()) {
//     // this is technically possible, but would make it possible to have nested transactions due to transactional becoming false
//     // nevertheless, using a sticky client inside a transaction is probably a mistake
//     throw new Error('Cannot use stickyClient inside a transaction');
//   }
//
//   using client = await useClient();
//   return await runWithStoredClient({ transactional: false, client }, fn);
// }
