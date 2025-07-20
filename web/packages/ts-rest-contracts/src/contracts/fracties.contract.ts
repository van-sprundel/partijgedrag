import { pageView } from '../schemas/utils.schema.js';
import { Motie } from '../schemas/moties.schema.js';
import { initContract } from '@ts-rest/core';
import { Fractie } from '../schemas/fracties.schema.js';

const c = initContract();

export const fractieContract = c.router(
  {
    getAll: pageView<Fractie>(c,'/'),
  },
  {
    pathPrefix: '/fracties',
  },
);
