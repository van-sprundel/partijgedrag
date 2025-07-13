// app/routes/dashboard-layout.tsx
import { Outlet, useLoaderData } from 'react-router';
import Navbar from '~/components/ui/navbar';
import { authLoader } from '~/loaders/auth-loader';
import type { Route } from '../../.react-router/types/app/+types/root';

export async function loader(loaderArgs: Route.LoaderArgs) {
  return await authLoader(loaderArgs);
}

export default function HomeLayout() {
  const user = useLoaderData();
  return (
    <>
        <Navbar />
        <Outlet />
    </>
  );
}