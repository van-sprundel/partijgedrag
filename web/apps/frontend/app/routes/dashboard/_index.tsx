import { initClient } from '@ts-rest/core';
import { authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import React, { useState } from 'react';
import { Outlet, redirect, useLoaderData } from 'react-router';
import { authLoader } from '~/loaders/auth-loader';

import type { Route } from '../+types/_index';

const client = initClient(authContract, {
  baseUrl: 'http://localhost:3000/api',
  baseHeaders: {
    'x-app-source': 'ts-rest',
  },
});

export async function loader(loaderData: Route.LoaderArgs) {
  return await authLoader(loaderData);
}

export default function Dashboard() {
  const user = useLoaderData();

  return (
    <>
      <div className="flex flex-col h-min-full w-full items-center justify-center ">
                    <h1>{
                      user?.email ?? 'Not logged in'
                    }</h1>
          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas viverra rutrum ante ut tempus. Nulla venenatis cursus turpis, et aliquet est fermentum sit amet. Interdum et malesuada fames ac ante ipsum primis in faucibus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Donec malesuada luctus sapien. Nullam sit amet felis orci. Integer ac bibendum augue. Proin interdum augue ex, sit amet vehicula neque viverra sit amet. Duis ullamcorper semper auctor. Aliquam vitae convallis lacus.

          Morbi commodo mollis mi non gravida. Quisque tincidunt sagittis felis ac suscipit. Vivamus facilisis in lacus eget euismod. Cras et commodo est, sit amet pulvinar arcu. Nullam blandit massa nec mauris accumsan tempus. Maecenas convallis quam sed lacinia vehicula. Nullam rhoncus diam eget purus tempus aliquam. Fusce cursus diam at ipsum tristique suscipit at et magna. Phasellus eu velit ullamcorper, maximus quam in, scelerisque justo. Proin sit amet rhoncus nulla. Mauris eget ante in risus bibendum sagittis. Ut ut scelerisque turpis. Aenean in lectus eget nisl aliquet mollis. Duis magna nibh, interdum sed leo vel, efficitur sodales nunc.

          Sed non condimentum felis. Sed semper, turpis in mollis tristique, tortor dolor luctus lacus, nec ultricies ligula quam vel tortor. Morbi luctus odio hendrerit, dignissim est vel, sagittis neque. Sed placerat mattis rhoncus. Fusce malesuada elit et libero posuere bibendum. Quisque condimentum ipsum at neque faucibus, id sodales lorem eleifend. Proin accumsan tincidunt faucibus. Sed ut condimentum sem, vel euismod purus. Fusce ornare mollis consectetur. Morbi nec tempor leo, id cursus diam. Sed sodales viverra sem, ac cursus nulla euismod eu. Morbi quis risus dictum, maximus lorem id, vulputate nisi. Nunc cursus semper sodales. Vivamus rutrum purus ut massa sodales, quis consectetur sapien consequat. Ut vel sapien porttitor, feugiat augue vitae, condimentum quam. Sed lacinia placerat ante, quis scelerisque mauris ornare eu.

          Nunc a feugiat magna. Vivamus varius urna sed eros auctor, a hendrerit augue vulputate. Nam ac neque a metus cursus hendrerit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Cras rutrum fringilla posuere. Ut diam massa, condimentum quis fringilla eu, cursus ac tortor. Etiam eu mi semper magna viverra porttitor. Duis pellentesque magna suscipit tristique mollis. Nam dignissim odio nec magna tristique viverra. Sed dapibus sapien convallis tristique laoreet. Mauris posuere pulvinar mi, at commodo sapien ornare vel. Curabitur ac rutrum lectus, vestibulum mattis eros. In et lectus eu nisl maximus blandit non pulvinar sem. Suspendisse feugiat porttitor velit. Aenean varius risus tortor, sed dignissim dui tincidunt in. Donec tellus lorem, tincidunt a euismod at, facilisis ultricies odio.

          Sed nibh mi, aliquam eget fermentum non, rutrum vitae libero. Quisque tempus fermentum arcu. Donec ut justo nisl. Fusce cursus id leo eu iaculis. Nulla tempor viverra est, ac venenatis elit ultricies id. Phasellus pellentesque rhoncus dolor, venenatis congue urna. Ut a faucibus ligula. Mauris eu metus nec nibh ultricies posuere ut eu tortor. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Sed sit amet tempus lectus.

          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas viverra rutrum ante ut tempus. Nulla venenatis cursus turpis, et aliquet est fermentum sit amet. Interdum et malesuada fames ac ante ipsum primis in faucibus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Donec malesuada luctus sapien. Nullam sit amet felis orci. Integer ac bibendum augue. Proin interdum augue ex, sit amet vehicula neque viverra sit amet. Duis ullamcorper semper auctor. Aliquam vitae convallis lacus.

          Morbi commodo mollis mi non gravida. Quisque tincidunt sagittis felis ac suscipit. Vivamus facilisis in lacus eget euismod. Cras et commodo est, sit amet pulvinar arcu. Nullam blandit massa nec mauris accumsan tempus. Maecenas convallis quam sed lacinia vehicula. Nullam rhoncus diam eget purus tempus aliquam. Fusce cursus diam at ipsum tristique suscipit at et magna. Phasellus eu velit ullamcorper, maximus quam in, scelerisque justo. Proin sit amet rhoncus nulla. Mauris eget ante in risus bibendum sagittis. Ut ut scelerisque turpis. Aenean in lectus eget nisl aliquet mollis. Duis magna nibh, interdum sed leo vel, efficitur sodales nunc.

          Sed non condimentum felis. Sed semper, turpis in mollis tristique, tortor dolor luctus lacus, nec ultricies ligula quam vel tortor. Morbi luctus odio hendrerit, dignissim est vel, sagittis neque. Sed placerat mattis rhoncus. Fusce malesuada elit et libero posuere bibendum. Quisque condimentum ipsum at neque faucibus, id sodales lorem eleifend. Proin accumsan tincidunt faucibus. Sed ut condimentum sem, vel euismod purus. Fusce ornare mollis consectetur. Morbi nec tempor leo, id cursus diam. Sed sodales viverra sem, ac cursus nulla euismod eu. Morbi quis risus dictum, maximus lorem id, vulputate nisi. Nunc cursus semper sodales. Vivamus rutrum purus ut massa sodales, quis consectetur sapien consequat. Ut vel sapien porttitor, feugiat augue vitae, condimentum quam. Sed lacinia placerat ante, quis scelerisque mauris ornare eu.

          Nunc a feugiat magna. Vivamus varius urna sed eros auctor, a hendrerit augue vulputate. Nam ac neque a metus cursus hendrerit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Cras rutrum fringilla posuere. Ut diam massa, condimentum quis fringilla eu, cursus ac tortor. Etiam eu mi semper magna viverra porttitor. Duis pellentesque magna suscipit tristique mollis. Nam dignissim odio nec magna tristique viverra. Sed dapibus sapien convallis tristique laoreet. Mauris posuere pulvinar mi, at commodo sapien ornare vel. Curabitur ac rutrum lectus, vestibulum mattis eros. In et lectus eu nisl maximus blandit non pulvinar sem. Suspendisse feugiat porttitor velit. Aenean varius risus tortor, sed dignissim dui tincidunt in. Donec tellus lorem, tincidunt a euismod at, facilisis ultricies odio.

          Sed nibh mi, aliquam eget fermentum non, rutrum vitae libero. Quisque tempus fermentum arcu. Donec ut justo nisl. Fusce cursus id leo eu iaculis. Nulla tempor viverra est, ac venenatis elit ultricies id. Phasellus pellentesque rhoncus dolor, venenatis congue urna. Ut a faucibus ligula. Mauris eu metus nec nibh ultricies posuere ut eu tortor. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Sed sit amet tempus lectus.


      </div>
    </>

    );
}