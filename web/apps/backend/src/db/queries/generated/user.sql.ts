/** Types generated for queries found in "src/db/queries/user.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

/** 'GetUnsafeUserByEmail' parameters type */
export interface IGetUnsafeUserByEmailParams {
  email: string;
}

/** 'GetUnsafeUserByEmail' return type */
export interface IGetUnsafeUserByEmailResult {
  email: string;
  id: number;
  password: string;
}

/** 'GetUnsafeUserByEmail' query type */
export interface IGetUnsafeUserByEmailQuery {
  params: IGetUnsafeUserByEmailParams;
  result: IGetUnsafeUserByEmailResult;
}

const getUnsafeUserByEmailIR: any = {"usedParamSet":{"email":true},"params":[{"name":"email","required":true,"transform":{"type":"scalar"},"locs":[{"a":34,"b":40}]}],"statement":"SELECT * from users WHERE email = :email!"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * from users WHERE email = :email!
 * ```
 */
export const getUnsafeUserByEmail = new PreparedQuery<IGetUnsafeUserByEmailParams,IGetUnsafeUserByEmailResult>(getUnsafeUserByEmailIR);


/** 'GetUserByEmail' parameters type */
export interface IGetUserByEmailParams {
  email: string;
}

/** 'GetUserByEmail' return type */
export interface IGetUserByEmailResult {
  email: string;
  id: number;
}

/** 'GetUserByEmail' query type */
export interface IGetUserByEmailQuery {
  params: IGetUserByEmailParams;
  result: IGetUserByEmailResult;
}

const getUserByEmailIR: any = {"usedParamSet":{"email":true},"params":[{"name":"email","required":true,"transform":{"type":"scalar"},"locs":[{"a":42,"b":48}]}],"statement":"SELECT id, email from users WHERE email = :email!"};

/**
 * Query generated from SQL:
 * ```
 * SELECT id, email from users WHERE email = :email!
 * ```
 */
export const getUserByEmail = new PreparedQuery<IGetUserByEmailParams,IGetUserByEmailResult>(getUserByEmailIR);


/** 'CreateUser' parameters type */
export interface ICreateUserParams {
  email: string;
  password: string;
}

/** 'CreateUser' return type */
export type ICreateUserResult = void;

/** 'CreateUser' query type */
export interface ICreateUserQuery {
  params: ICreateUserParams;
  result: ICreateUserResult;
}

const createUserIR: any = {"usedParamSet":{"email":true,"password":true},"params":[{"name":"email","required":true,"transform":{"type":"scalar"},"locs":[{"a":48,"b":54}]},{"name":"password","required":true,"transform":{"type":"scalar"},"locs":[{"a":57,"b":66}]}],"statement":"INSERT INTO users (\"email\",  \"password\")\nVALUES(:email!, :password!)"};

/**
 * Query generated from SQL:
 * ```
 * INSERT INTO users ("email",  "password")
 * VALUES(:email!, :password!)
 * ```
 */
export const createUser = new PreparedQuery<ICreateUserParams,ICreateUserResult>(createUserIR);


/** 'UpdatePassword' parameters type */
export interface IUpdatePasswordParams {
  id: number;
  password: string;
}

/** 'UpdatePassword' return type */
export type IUpdatePasswordResult = void;

/** 'UpdatePassword' query type */
export interface IUpdatePasswordQuery {
  params: IUpdatePasswordParams;
  result: IUpdatePasswordResult;
}

const updatePasswordIR: any = {"usedParamSet":{"password":true,"id":true},"params":[{"name":"password","required":true,"transform":{"type":"scalar"},"locs":[{"a":30,"b":39}]},{"name":"id","required":true,"transform":{"type":"scalar"},"locs":[{"a":52,"b":55}]}],"statement":"UPDATE users\nSET \"password\" = :password!\nWHERE id = :id!"};

/**
 * Query generated from SQL:
 * ```
 * UPDATE users
 * SET "password" = :password!
 * WHERE id = :id!
 * ```
 */
export const updatePassword = new PreparedQuery<IUpdatePasswordParams,IUpdatePasswordResult>(updatePasswordIR);


