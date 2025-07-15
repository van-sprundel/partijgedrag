// app/routes/dashboard-layout.tsx
import { Outlet, useLoaderData } from 'react-router';

export default function HomeLayout() {
  return (
    <>

        <Outlet />
    </>
  );
}