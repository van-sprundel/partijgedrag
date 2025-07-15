import { initContract } from '@ts-rest/core';
import { Motie } from '../schemas/moties.schema.js';

const c = initContract();

// TODO fix typing and automatic documentation
export const motieContract = c.router(
  {
    getAll: {
      method: 'GET',
      path: '/',
      responses: {
        200: c.type<{ moties: Motie[] }>(),
      },
    },
  },
  {
    pathPrefix: '/moties',
  },
);