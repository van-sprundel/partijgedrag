import { index, layout, prefix, type RouteConfig, type RouteConfigEntry } from '@react-router/dev/routes';
import { flatRoutes } from '@react-router/fs-routes';
import type { RouteObject } from 'react-router';

export const authRoutes = [...await flatRoutes({
  rootDirectory: 'routes/auth',
})] satisfies RouteConfig;


export const dashboardRoutes = [
  ...await flatRoutes({
    rootDirectory: 'routes/dashboard'
  }),
  ...prefix("sub-route",
  [...await flatRoutes({
    rootDirectory: 'routes/dashboard/sub-route'
  }), ...prefix("sub-sub-route", [
  ...await flatRoutes({
    rootDirectory: 'routes/dashboard/sub-route/sub-sub-route'
  })])])
]



export default [
  layout('layouts/home-layout.tsx',
    [
      index('routes/_index.tsx'),
      ...prefix("auth", authRoutes)]
    ),
  layout('layouts/dashboard-layout.tsx',
    [
      ...prefix("dashboard", dashboardRoutes),
      ]
    )
]

