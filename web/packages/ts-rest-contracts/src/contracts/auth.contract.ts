import { initContract } from '@ts-rest/core';
import {
  forgotPasswordParamSchema, forgotPasswordQueryParamSchema,
  loginBodySchema, refreshTokenBodySchema,
  registerBodySchema,
  resetPasswordBodySchema,
  resetPasswordParamSchema,
} from '../schemas/auth.schema.js';

const c = initContract();

// TODO fix typing and automatic documentation
export const authContract = c.router(
  {
    register: {
      method: 'POST',
      path: '/register',
      body: registerBodySchema,
      responses: {
        200: c.type<{ message: string }>(),
      },
    },
    login: {
      method: 'POST',
      path: '/login',
      body: loginBodySchema,
      responses: {
        200: c.type<{ message: string}>(),
        401: c.type<{ message: string, inputErrors: Record<keyof typeof loginBodySchema, string[]>}>(),
        500: c.type<{ message: string}>(),

      },
      strictStatusCodes: true,
    },
    logout: {
      method: 'GET',
      path: '/logout',
      responses: {
        200: c.type<{ message: string }>(),
      },
    },
    me: {
      method: 'GET',
      path: '/me',
      responses: {
        200: c.type<{ id: number; email: string }>(),
      },
    },
    forgotPassword: {
      method: 'GET',
      path: '/forgot-password/:email',
      pathParams: forgotPasswordParamSchema,
      query: forgotPasswordQueryParamSchema,
      responses: {
        200: c.type<{ message: string }>(),
      },
    },
    passwordReset: {
      method: 'POST',
      path: '/reset-password/:token',
      pathParams: resetPasswordParamSchema,
      body: resetPasswordBodySchema,
      responses: {
        200: c.type<{ message: string }>(),
      },
    },
      refreshToken: {
        method: 'POST',
        path: '/refresh-token',
        body: refreshTokenBodySchema,
        responses: {
          200: c.type<{ accessToken: string, refreshToken: string }>(),
        },
      },
  },
  {
    pathPrefix: '/auth',
  },
);
