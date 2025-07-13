import { appContract, authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { initClient, type ServerInferResponses } from '@ts-rest/core';
import {fractieContract} from "@fullstack-typescript-template/ts-rest-contracts/dist/contracts/fracties.contract";

// export const API_BASE_URL = `${process.env.HOST}:${process.env.FRONTEND_PORT}/api`;
export const API_BASE_URL = 'http://localhost:3000/api'
// TODO put in vite
export const baseClient = initClient(appContract, {
  baseUrl: API_BASE_URL,
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
  withCredentials: true,
});

export type BaseClientResponseShapes = ServerInferResponses<typeof appContract, 200>;

export type AuthClientResponseShapes = ServerInferResponses<typeof authContract.login>;

export const authClient = initClient(authContract, {
  baseUrl: API_BASE_URL,
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
  withCredentials: true,
});

export const fractiesClient = initClient(fractieContract,  {
    baseUrl: API_BASE_URL,
})