// import { PreparedQuery, sql as sqlTaggedTemplate } from '@pgtyped/runtime';
// import { useClient } from './client.js';
// import * as user from './queries/generated/user.sql.js';
// import * as userSession from './queries/generated/user-session.sql.js';
// import * as passwordResetRequest from './queries/generated/password-request.sql.js';
// import dotenv from 'dotenv';
//
// dotenv.config();
//
// /**
//  * Object through which SQL queries in /services/db/queries are exposed.
//  */
// export const db = {
//   user: wrapPreparedQueries(user),
//   userSession: wrapPreparedQueries(userSession),
//   passwordResetRequest: wrapPreparedQueries(passwordResetRequest),
// };
//
// /**
//  * Sql tagged template that automatically uses the correct client.
//  */
// export const sql = <T extends { params: any; result: any }>(
//   arr: TemplateStringsArray,
// ) => {
//   const preparedQuery = sqlTaggedTemplate<T>(arr);
//   return async (param: T['params']): Promise<T['result'][]> => {
//     using client = await useClient();
//     return preparedQuery.run(param, client);
//   };
// };
//
// /**
//  * Wrap an object of prepared queries such that their client is automatically provided.
//  */
// function wrapPreparedQueries<const T extends PreparedQueries>(
//   queries: T,
// ): WrappedPreparedQueries<T> {
//   return Object.fromEntries(
//     Object.entries(queries).map(([queryName, query]) => [
//       queryName,
//       wrapPreparedQuery(query),
//     ]),
//   ) as WrappedPreparedQueries<T>;
// }
//
// function wrapPreparedQuery<Param, Result>(
//   query: PreparedQuery<Param, Result>,
// ): (param: Param) => Promise<Result[]> {
//   return async (param) => {
//     using client = await useClient();
//     return await query.run(param, client);
//   };
// }
//
// type PreparedQueries = {
//   [queryName: string]: PreparedQuery<any, any>;
// };
//
// type WrappedPreparedQueries<T extends PreparedQueries> = {
//   [queryName in keyof T]: WrappedPreparedQuery<T[queryName]>;
// };
//
// type WrappedPreparedQuery<P extends PreparedQuery<any, any>> =
//   P extends PreparedQuery<infer Param, infer Result>
//     ? (param: Param) => Promise<Result[]>
//     : never;
