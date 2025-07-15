import { initContract } from '@ts-rest/core';
import { Motie } from '../schemas/moties.schema.js';
import { z } from 'zod';
import { pageSchema, pageView } from '../schemas/utils.schema.js';

const c = initContract();


// TODO fix typing and automatic documentation
export const motieContract = c.router(
  {
    getAll: pageView<Motie>(c,'/'),
  },
  {
    pathPrefix: '/moties',
  },
);