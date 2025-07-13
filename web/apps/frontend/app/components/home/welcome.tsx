import { appContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { initClient } from '@ts-rest/core';
import { Button } from '~/components/ui/button';
import { Link } from 'react-router';
import type { BaseClientResponseShapes } from '~/clients';

export const client = initClient(appContract, {
  baseUrl: '/api',
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
});


export function Welcome({ appInfo }: { appInfo: BaseClientResponseShapes['getAppInfo']['body'] }) {

  return (
    <div className="relative pt-32 pb-20 sm:pt-40 sm:pb-24 ">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h1 className="text-4xl sm:text-6xl lg:text-7xl font-bold tracking-tight mb-8 ">
          {appInfo.name}: v{appInfo.version}
          <br />
          <span className="bg-gradient-to-br from-primary to-primary/40 bg-clip-text text-transparent" >for your teams</span>
        </h1>
        <p className="max-w-2xl mx-auto text-lg sm:text-xl text-gray-400 mb-10">
          A full-stack starter template for your teams to build fast, modern, and beautiful web applications.
        </p>
        <Button className="relative group px-8 py-6 text-lg bg-gradient-to-r from-primary to-primary/40 hover:opacity-90" >
          <Link to="/auth/login">Get Started</Link>
        </Button>
      </div>
    </div>
  );
}


