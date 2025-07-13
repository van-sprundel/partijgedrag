import { initContract } from '@ts-rest/core';
import { appContract } from './app.contract.js';
import { authContract } from './auth.contract.js';
import {fractieContract} from "./fracties.contract.js";

const c = initContract();

export const rootContract = c.router({
  App: appContract,
  Auth: authContract,
  Fractie: fractieContract
});
