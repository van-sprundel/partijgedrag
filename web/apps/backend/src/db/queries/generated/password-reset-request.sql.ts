/** Types generated for queries found in "src/db/queries/password-reset-request.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

export type DateOrString = Date | string;

/** 'GetPasswordResetRequestByUserId' parameters type */
export interface IGetPasswordResetRequestByUserIdParams {
  userId?: number | null | void;
}

/** 'GetPasswordResetRequestByUserId' return type */
export interface IGetPasswordResetRequestByUserIdResult {
  expiresAt: Date;
  id: number;
  token: string;
  userId: number;
}

/** 'GetPasswordResetRequestByUserId' query type */
export interface IGetPasswordResetRequestByUserIdQuery {
  params: IGetPasswordResetRequestByUserIdParams;
  result: IGetPasswordResetRequestByUserIdResult;
}

const getPasswordResetRequestByUserIdIR: any = {"usedParamSet":{"userId":true},"params":[{"name":"userId","required":false,"transform":{"type":"scalar"},"locs":[{"a":55,"b":61}]}],"statement":"SELECT * FROM password_reset_requests WHERE \"userId\" = :userId"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM password_reset_requests WHERE "userId" = :userId
 * ```
 */
export const getPasswordResetRequestByUserId = new PreparedQuery<IGetPasswordResetRequestByUserIdParams,IGetPasswordResetRequestByUserIdResult>(getPasswordResetRequestByUserIdIR);


/** 'CreatePasswordResetRequest' parameters type */
export interface ICreatePasswordResetRequestParams {
  expiresAt: DateOrString;
  token: string;
  userId: number;
}

/** 'CreatePasswordResetRequest' return type */
export interface ICreatePasswordResetRequestResult {
  expiresAt: Date;
  id: number;
  token: string;
  userId: number;
}

/** 'CreatePasswordResetRequest' query type */
export interface ICreatePasswordResetRequestQuery {
  params: ICreatePasswordResetRequestParams;
  result: ICreatePasswordResetRequestResult;
}

const createPasswordResetRequestIR: any = {"usedParamSet":{"userId":true,"token":true,"expiresAt":true},"params":[{"name":"userId","required":true,"transform":{"type":"scalar"},"locs":[{"a":77,"b":84}]},{"name":"token","required":true,"transform":{"type":"scalar"},"locs":[{"a":87,"b":93}]},{"name":"expiresAt","required":true,"transform":{"type":"scalar"},"locs":[{"a":96,"b":106}]}],"statement":"INSERT INTO password_reset_requests (\"userId\", \"token\", \"expiresAt\") VALUES (:userId!, :token!, :expiresAt!)\nRETURNING *"};

/**
 * Query generated from SQL:
 * ```
 * INSERT INTO password_reset_requests ("userId", "token", "expiresAt") VALUES (:userId!, :token!, :expiresAt!)
 * RETURNING *
 * ```
 */
export const createPasswordResetRequest = new PreparedQuery<ICreatePasswordResetRequestParams,ICreatePasswordResetRequestResult>(createPasswordResetRequestIR);


/** 'DeletePasswordResetRequestByUserId' parameters type */
export interface IDeletePasswordResetRequestByUserIdParams {
  userId: number;
}

/** 'DeletePasswordResetRequestByUserId' return type */
export type IDeletePasswordResetRequestByUserIdResult = void;

/** 'DeletePasswordResetRequestByUserId' query type */
export interface IDeletePasswordResetRequestByUserIdQuery {
  params: IDeletePasswordResetRequestByUserIdParams;
  result: IDeletePasswordResetRequestByUserIdResult;
}

const deletePasswordResetRequestByUserIdIR: any = {"usedParamSet":{"userId":true},"params":[{"name":"userId","required":true,"transform":{"type":"scalar"},"locs":[{"a":53,"b":60}]}],"statement":"DELETE FROM password_reset_requests WHERE \"userId\" = :userId!"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM password_reset_requests WHERE "userId" = :userId!
 * ```
 */
export const deletePasswordResetRequestByUserId = new PreparedQuery<IDeletePasswordResetRequestByUserIdParams,IDeletePasswordResetRequestByUserIdResult>(deletePasswordResetRequestByUserIdIR);


/** 'DeleteExpiredPasswordResetRequests' parameters type */
export type IDeleteExpiredPasswordResetRequestsParams = void;

/** 'DeleteExpiredPasswordResetRequests' return type */
export type IDeleteExpiredPasswordResetRequestsResult = void;

/** 'DeleteExpiredPasswordResetRequests' query type */
export interface IDeleteExpiredPasswordResetRequestsQuery {
  params: IDeleteExpiredPasswordResetRequestsParams;
  result: IDeleteExpiredPasswordResetRequestsResult;
}

const deleteExpiredPasswordResetRequestsIR: any = {"usedParamSet":{},"params":[],"statement":"DELETE FROM password_reset_requests WHERE \"expiresAt\" < now()"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM password_reset_requests WHERE "expiresAt" < now()
 * ```
 */
export const deleteExpiredPasswordResetRequests = new PreparedQuery<IDeleteExpiredPasswordResetRequestsParams,IDeleteExpiredPasswordResetRequestsResult>(deleteExpiredPasswordResetRequestsIR);


/** 'GetPasswordResetRequestByToken' parameters type */
export interface IGetPasswordResetRequestByTokenParams {
  token?: string | null | void;
}

/** 'GetPasswordResetRequestByToken' return type */
export interface IGetPasswordResetRequestByTokenResult {
  expiresAt: Date;
  id: number;
  token: string;
  userId: number;
}

/** 'GetPasswordResetRequestByToken' query type */
export interface IGetPasswordResetRequestByTokenQuery {
  params: IGetPasswordResetRequestByTokenParams;
  result: IGetPasswordResetRequestByTokenResult;
}

const getPasswordResetRequestByTokenIR: any = {"usedParamSet":{"token":true},"params":[{"name":"token","required":false,"transform":{"type":"scalar"},"locs":[{"a":54,"b":59}]}],"statement":"SELECT * FROM password_reset_requests WHERE \"token\" = :token"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM password_reset_requests WHERE "token" = :token
 * ```
 */
export const getPasswordResetRequestByToken = new PreparedQuery<IGetPasswordResetRequestByTokenParams,IGetPasswordResetRequestByTokenResult>(getPasswordResetRequestByTokenIR);


/** 'DeletePasswordResetRequestByToken' parameters type */
export interface IDeletePasswordResetRequestByTokenParams {
  token: string;
}

/** 'DeletePasswordResetRequestByToken' return type */
export type IDeletePasswordResetRequestByTokenResult = void;

/** 'DeletePasswordResetRequestByToken' query type */
export interface IDeletePasswordResetRequestByTokenQuery {
  params: IDeletePasswordResetRequestByTokenParams;
  result: IDeletePasswordResetRequestByTokenResult;
}

const deletePasswordResetRequestByTokenIR: any = {"usedParamSet":{"token":true},"params":[{"name":"token","required":true,"transform":{"type":"scalar"},"locs":[{"a":52,"b":58}]}],"statement":"DELETE FROM password_reset_requests WHERE \"token\" = :token!"};

/**
 * Query generated from SQL:
 * ```
 * DELETE FROM password_reset_requests WHERE "token" = :token!
 * ```
 */
export const deletePasswordResetRequestByToken = new PreparedQuery<IDeletePasswordResetRequestByTokenParams,IDeletePasswordResetRequestByTokenResult>(deletePasswordResetRequestByTokenIR);


