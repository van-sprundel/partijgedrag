import { redirect} from 'react-router';
import { authClient } from '~/clients';
import type { Route } from '../+types/_index';

export async function loader({request}: Route.LoaderArgs) {
  const cookieHeader = request.headers.get('cookie');
  if (cookieHeader) {
    const { body, status } = await authClient.logout({
      extraHeaders: {
        cookie: cookieHeader,
      }
    });
    if (status !== 200) {
      return;
    }
    return redirect('/', {
      headers: {
        'Set-Cookie': 'auth=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT',
      },
    });
  }
  return redirect('/');
}