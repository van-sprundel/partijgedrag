import { initContract } from '@ts-rest/core';
import {Fractie} from "../schemas/fracties.schema.js";

const c = initContract();

// TODO fix typing and automatic documentation
export const fractieContract = c.router(
  {
    getAll: {
      method: 'GET',
      path: '/',
      responses: {
        200: c.type<{ fracties: Fractie[] }>(),
      },
    },
  },
  {
    pathPrefix: '/fracties',
  },
);
