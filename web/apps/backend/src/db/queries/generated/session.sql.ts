/** Types generated for queries found in "src/database/queries/user-session.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

/** Query 'GetSessionById' is invalid, so its result is assigned type 'never'.
 *  */
export type IGetSessionByIdResult = never;

/** Query 'GetSessionById' is invalid, so its parameters are assigned type 'never'.
 *  */
export type IGetSessionByIdParams = never;

const getSessionByIdIR: any = {"usedParamSet":{"id":true},"params":[{"name":"id","required":false,"transform":{"type":"scalar"},"locs":[{"a":37,"b":39}]}],"statement":"SELECT * FROM \"Session\" WHERE \"id\" = :id"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM "Session" WHERE "id" = :id
 * ```
 */
export const getSessionById = new PreparedQuery<IGetSessionByIdParams,IGetSessionByIdResult>(getSessionByIdIR);


