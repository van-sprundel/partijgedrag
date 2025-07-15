import { initContract } from '@ts-rest/core';
import { Motie } from '../schemas/moties.schema.js';
import { z } from 'zod';
import { pageSchema } from '../schemas/utils.schema.js';

const c = initContract();

// TODO fix typing and automatic documentation
export const motieContract = c.router(
  {
    getAll: {
      method: 'GET',
      path: '/',
      query: pageSchema,
      responses: {
        200: c.type<{ moties: Motie[] }>(),
      },
    },
  },
  {
    pathPrefix: '/moties',
  },
);