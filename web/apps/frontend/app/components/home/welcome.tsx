import { appContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { initClient } from '@ts-rest/core';
import MotionCard from '~/components/motion';

export const client = initClient(appContract, {
  baseUrl: '/api',
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
});


export function Welcome() {

  return (
      <MotionCard />
  );
}


