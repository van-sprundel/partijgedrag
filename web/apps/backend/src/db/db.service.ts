import { Injectable } from '@nestjs/common';
import pg from 'pg';
import { PreparedQuery, sql as sqlTaggedTemplate } from '@pgtyped/runtime';
import { generatedQueries, ServiceWithGeneratedQueries, WrappedPreparedQueries } from './generated-queries.js';
import dotenv from 'dotenv';
import { AsyncLocalStorage } from 'async_hooks';

dotenv.config();

type ManagedPoolClient = Omit<pg.PoolClient, 'release'>;

type DisposablePoolClient = ManagedPoolClient & Disposable;

type StoredClient = {
  transactional: boolean;
  client: DisposablePoolClient;
};

type PreparedQueries = {
  [queryName: string]: PreparedQuery<any, any>;
};

@Injectable()
export class DbService extends ServiceWithGeneratedQueries {
  private readonly pool = new pg.Pool({
    connectionString: process.env.DATABASE_URL,
    password: process.env.DATABASE_PASSWORD,
  });

  constructor() {
    super();
    for (const [name, queries] of Object.entries(generatedQueries)) {
      // @ts-expect-error - ignore generated queries assigned to the db property
      this[name] = this.wrapPreparedQueries(queries);
    }
  }

  private sql<T extends { params: any; result: any }>(arr: TemplateStringsArray) {
    const preparedQuery = sqlTaggedTemplate<T>(arr);
    return async (param: T['params']): Promise<T['result'][]> => {
      using client = await this.useClient();
      return preparedQuery.run(param, client);
    };
  }

  private asyncLocalStorage = new AsyncLocalStorage<StoredClient>();

  private wrapPreparedQueries<const T extends PreparedQueries>(queries: T): WrappedPreparedQueries<T> {
    return Object.fromEntries(
      Object.entries(queries).map(([queryName, query]) => [queryName, this.wrapPreparedQuery(query)]),
    ) as WrappedPreparedQueries<T>;
  }

  private wrapPreparedQuery<Param, Result>(query: PreparedQuery<Param, Result>): (param: Param) => Promise<Result[]> {
    return async (param) => {
      using client = await this.useClient();
      return await query.run(param, client);
    };
  }

  private async useClient(): Promise<DisposablePoolClient> {
    const storedClient = this.getStoredClient();

    if (storedClient) {
      const dispose = storedClient.client[Symbol.dispose];

      storedClient.client[Symbol.dispose] = () => {
        storedClient.client[Symbol.dispose] = dispose;
      };

      return storedClient.client;
    }

    const client = await this.pool.connect();
    let released = false;
    return Object.assign(client, {
      [Symbol.dispose]: () => {
        if (released) return;
        client.release();
        released = true;
      },
    });
  }

  private getStoredClient(): StoredClient | undefined {
    return this.asyncLocalStorage.getStore();
  }

  private runWithStoredClient<T>(storedClient: StoredClient, fn: () => T): T {
    return this.asyncLocalStorage.run(storedClient, fn);
  }

  private async stickyClient<T>(fn: () => Promise<T>): Promise<T> {
    if (this.inTransaction()) {
      throw new Error('Cannot use stickyClient inside a transaction');
    }

    using client = await this.useClient();
    return await this.runWithStoredClient({ transactional: false, client }, fn);
  }

  private inTransaction(): boolean {
    return this.getStoredClient()?.transactional === true;
  }
}

const db = new DbService();
