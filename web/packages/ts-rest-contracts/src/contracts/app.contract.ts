import { initContract } from '@ts-rest/core';

const c = initContract();

export const appContract = c.router({
  getAppInfo: {
    method: 'GET',
    path: '/app-info',
    responses: {
      200: c.type<{ version: string; name: string }>(),
    },
  },
});
