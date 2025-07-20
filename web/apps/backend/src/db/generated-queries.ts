/**
 * AUTO-GENERATED FILE — DO NOT EDIT MANUALLY
 * Run "npm run generate:queries" to regenerate.
 */

import { PreparedQuery } from '@pgtyped/runtime';
import * as fracties from './queries/generated/fracties.sql.js';
import * as moties from './queries/generated/moties.sql.js';
import * as passwordResetRequest from './queries/generated/password-reset-request.sql.js';
import * as session from './queries/generated/session.sql.js';
import * as userSession from './queries/generated/user-session.sql.js';
import * as user from './queries/generated/user.sql.js';

export type WrappedPreparedQuery<P extends PreparedQuery<any, any>> =
  P extends PreparedQuery<infer Param, infer Result> ? (param: Param) => Promise<Result[]> : never;

export type WrappedPreparedQueries<T> = {
  [K in keyof T]: T[K] extends PreparedQuery<any, any> ? WrappedPreparedQuery<T[K]> : never;
};

export type GeneratedQueryTypes = {
  fracties: typeof fracties;
  moties: typeof moties;
  passwordResetRequest: typeof passwordResetRequest;
  session: typeof session;
  userSession: typeof userSession;
  user: typeof user;
};

export abstract class ServiceWithGeneratedQueries {
  declare fracties: WrappedPreparedQueries<GeneratedQueryTypes['fracties']>;
  declare moties: WrappedPreparedQueries<GeneratedQueryTypes['moties']>;
  declare passwordResetRequest: WrappedPreparedQueries<GeneratedQueryTypes['passwordResetRequest']>;
  declare session: WrappedPreparedQueries<GeneratedQueryTypes['session']>;
  declare userSession: WrappedPreparedQueries<GeneratedQueryTypes['userSession']>;
  declare user: WrappedPreparedQueries<GeneratedQueryTypes['user']>;
}

export const generatedQueries: GeneratedQueryTypes = {
  fracties,
  moties,
  passwordResetRequest,
  session,
  userSession,
  user,
};
