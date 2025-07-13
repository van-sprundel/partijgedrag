
import { authClient } from '~/clients';
import { redirect } from 'react-router';
import type { Route } from '../../.react-router/types/app/+types/root';

export async function authLoader(loaderArgs: Route.LoaderArgs) {
  const { request } = loaderArgs;
  const cookieHeader = request.headers.get('cookie'); // ✅ Get session from cookies
  if (cookieHeader) {
    const { body, status } = await authClient.me({
      extraHeaders: {
        cookie: cookieHeader,
      }
    });
    if (status !== 200) {
      console.log('Error auth guard')
      console.log({status, body})
      return null;
    }
    return body;
  }
}

export async function nonAuthLoader(loaderArgs: Route.LoaderArgs) {
  const user = await authLoader(loaderArgs  );
  if (user) {
    console.log('redirecting')
    return redirect( '/');
  }
  return
}