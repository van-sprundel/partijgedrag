/** Types generated for queries found in "src/db/queries/user-session.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

export type DateOrString = Date | string;

/** 'GetSessionById' parameters type */
export interface IGetSessionByIdParams {
  id?: number | null | void;
}

/** 'GetSessionById' return type */
export interface IGetSessionByIdResult {
  accessToken: string;
  expiresAt: Date;
  id: number;
  refreshToken: string;
  userId: number;
}

/** 'GetSessionById' query type */
export interface IGetSessionByIdQuery {
  params: IGetSessionByIdParams;
  result: IGetSessionByIdResult;
}

const getSessionByIdIR: any = {"usedParamSet":{"id":true},"params":[{"name":"id","required":false,"transform":{"type":"scalar"},"locs":[{"a":41,"b":43}]}],"statement":"SELECT * FROM user_sessions WHERE \"id\" = :id"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM user_sessions WHERE "id" = :id
 * ```
 */
export const getSessionById = new PreparedQuery<IGetSessionByIdParams,IGetSessionByIdResult>(getSessionByIdIR);


/** 'GetSessionByAccessToken' parameters type */
export interface IGetSessionByAccessTokenParams {
  accessToken?: string | null | void;
}

/** 'GetSessionByAccessToken' return type */
export interface IGetSessionByAccessTokenResult {
  accessToken: string;
  expiresAt: Date;
  id: number;
  refreshToken: string;
  userId: number;
}

/** 'GetSessionByAccessToken' query type */
export interface IGetSessionByAccessTokenQuery {
  params: IGetSessionByAccessTokenParams;
  result: IGetSessionByAccessTokenResult;
}

const getSessionByAccessTokenIR: any = {"usedParamSet":{"accessToken":true},"params":[{"name":"accessToken","required":false,"transform":{"type":"scalar"},"locs":[{"a":50,"b":61}]}],"statement":"SELECT * FROM user_sessions WHERE \"accessToken\" = :accessToken"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM user_sessions WHERE "accessToken" = :accessToken
 * ```
 */
export const getSessionByAccessToken = new PreparedQuery<IGetSessionByAccessTokenParams,IGetSessionByAccessTokenResult>(getSessionByAccessTokenIR);


/** 'GetSessionByRefreshToken' parameters type */
export interface IGetSessionByRefreshTokenParams {
  refreshToken?: string | null | void;
}

/** 'GetSessionByRefreshToken' return type */
export interface IGetSessionByRefreshTokenResult {
  accessToken: string;
  expiresAt: Date;
  id: number;
  refreshToken: string;
  userId: number;
}

/** 'GetSessionByRefreshToken' query type */
export interface IGetSessionByRefreshTokenQuery {
  params: IGetSessionByRefreshTokenParams;
  result: IGetSessionByRefreshTokenResult;
}

const getSessionByRefreshTokenIR: any = {"usedParamSet":{"refreshToken":true},"params":[{"name":"refreshToken","required":false,"transform":{"type":"scalar"},"locs":[{"a":51,"b":63}]}],"statement":"SELECT * FROM user_sessions WHERE \"refreshToken\" = :refreshToken"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM user_sessions WHERE "refreshToken" = :refreshToken
 * ```
 */
export const getSessionByRefreshToken = new PreparedQuery<IGetSessionByRefreshTokenParams,IGetSessionByRefreshTokenResult>(getSessionByRefreshTokenIR);


/** 'CreateSession' parameters type */
export interface ICreateSessionParams {
  accessToken: string;
  expiresAt: DateOrString;
  refreshToken: string;
  userId: number;
}

/** 'CreateSession' return type */
export interface ICreateSessionResult {
  accessToken: string;
  expiresAt: Date;
  id: number;
  refreshToken: string;
  userId: number;
}

/** 'CreateSession' query type */
export interface ICreateSessionQuery {
  params: ICreateSessionParams;
  result: ICreateSessionResult;
}

const createSessionIR: any = {"usedParamSet":{"userId":true,"expiresAt":true,"accessToken":true,"refreshToken":true},"params":[{"name":"userId","required":true,"transform":{"type":"scalar"},"locs":[{"a":89,"b":96}]},{"name":"expiresAt","required":true,"transform":{"type":"scalar"},"locs":[{"a":99,"b":109}]},{"name":"accessToken","required":true,"transform":{"type":"scalar"},"locs":[{"a":112,"b":124}]},{"name":"refreshToken","required":true,"transform":{"type":"scalar"},"locs":[{"a":127,"b":140}]}],"statement":"INSERT INTO user_sessions (\"userId\", \"expiresAt\", \"accessToken\", \"refreshToken\")\nVALUES (:userId!, :expiresAt!, :accessToken!, :refreshToken!)\nRETURNING *"};

/**
 * Query generated from SQL:
 * ```
 * INSERT INTO user_sessions ("userId", "expiresAt", "accessToken", "refreshToken")
 * VALUES (:userId!, :expiresAt!, :accessToken!, :refreshToken!)
 * RETURNING *
 * ```
 */
export const createSession = new PreparedQuery<ICreateSessionParams,ICreateSessionResult>(createSessionIR);


/** 'DeleteSessionByAccessToken' parameters type */
export interface IDeleteSessionByAccessTokenParams {
  accessToken: string;
}

/** 'DeleteSessionByAccessToken' return type */
export type IDeleteSessionByAccessTokenResult = void;

/** 'DeleteSessionByAccessToken' query type */
export interface IDeleteSessionByAccessTokenQuery {
  params: IDeleteSessionByAccessTokenParams;
  result: IDeleteSessionByAccessTokenResult;
}

const deleteSessionByAccessTokenIR: any = {"usedParamSet":{"accessToken":true},"params":[{"name":"accessToken","required":true,"transform":{"type":"scalar"},"locs":[{"a":48,"b":60}]}],"statement":"DELETE FROM user_sessions WHERE \"accessToken\" = :accessToken!"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM user_sessions WHERE "accessToken" = :accessToken!
 * ```
 */
export const deleteSessionByAccessToken = new PreparedQuery<IDeleteSessionByAccessTokenParams,IDeleteSessionByAccessTokenResult>(deleteSessionByAccessTokenIR);


/** 'DeleteAllSessionsOfUser' parameters type */
export interface IDeleteAllSessionsOfUserParams {
  userId: number;
}

/** 'DeleteAllSessionsOfUser' return type */
export type IDeleteAllSessionsOfUserResult = void;

/** 'DeleteAllSessionsOfUser' query type */
export interface IDeleteAllSessionsOfUserQuery {
  params: IDeleteAllSessionsOfUserParams;
  result: IDeleteAllSessionsOfUserResult;
}

const deleteAllSessionsOfUserIR: any = {"usedParamSet":{"userId":true},"params":[{"name":"userId","required":true,"transform":{"type":"scalar"},"locs":[{"a":43,"b":50}]}],"statement":"DELETE FROM user_sessions WHERE \"userId\" = :userId!"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM user_sessions WHERE "userId" = :userId!
 * ```
 */
export const deleteAllSessionsOfUser = new PreparedQuery<IDeleteAllSessionsOfUserParams,IDeleteAllSessionsOfUserResult>(deleteAllSessionsOfUserIR);


/** 'DeleteExpiredSessions' parameters type */
export type IDeleteExpiredSessionsParams = void;

/** 'DeleteExpiredSessions' return type */
export type IDeleteExpiredSessionsResult = void;

/** 'DeleteExpiredSessions' query type */
export interface IDeleteExpiredSessionsQuery {
  params: IDeleteExpiredSessionsParams;
  result: IDeleteExpiredSessionsResult;
}

const deleteExpiredSessionsIR: any = {"usedParamSet":{},"params":[],"statement":"DELETE FROM user_sessions WHERE \"expiresAt\" < now()"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM user_sessions WHERE "expiresAt" < now()
 * ```
 */
export const deleteExpiredSessions = new PreparedQuery<IDeleteExpiredSessionsParams,IDeleteExpiredSessionsResult>(deleteExpiredSessionsIR);


