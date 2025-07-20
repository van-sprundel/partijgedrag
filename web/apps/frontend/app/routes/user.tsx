import type { Route } from '../../.react-router/types/app/+types/root';
import { initClient } from '@ts-rest/core';
import { authContract } from '@fullstack-typescript-template/ts-rest-contracts';

const client = initClient(authContract, {
  baseUrl: 'http://localhost:3000/api',
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
  credentials: 'include',
});

export async function loader({ request }: Route.LoaderArgs) {
  const me = await client.me({
    fetchOptions: {
      credentials: 'include', // ✅ Force cookies to be included
    },
  });
  return me;
}