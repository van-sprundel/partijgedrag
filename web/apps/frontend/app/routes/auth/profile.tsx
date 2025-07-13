import Profile from "~/components/auth/profile";

import { useLoaderData } from 'react-router';
import { authLoader } from '~/loaders/auth-loader';
import type { Route } from '../../../.react-router/types/app/+types/root';

async function loader(loaderData: Route.LoaderArgs) {
  return await authLoader(loaderData);
}

export default function Page() {
  const data = useLoaderData();

  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm text-center">
          <Profile user={data} />
      </div>
    </div>
  );
}
