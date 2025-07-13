import type { Route } from "../+types/root";
import MouseMoveEffect from '~/components/effects/mouse-move-effect';
import { authLoader } from '~/loaders/auth-loader';
import { useLoaderData } from 'react-router';
import {baseClient, fractiesClient} from '~/clients';
import { Welcome } from '~/components/home/welcome';
import {useEffect} from "react";

export function meta({}: Route.MetaArgs) {
  return [
    { title: 'Fullstack TypeScript Template' },
    { name: 'description', content: 'A fullstack template for TypeScript for creating production-ready applications.' },
  ];
}

export async function loader({ request }: Route.LoaderArgs) {
  const { body, status } = await baseClient.getAppInfo();
  if (status !== 200) {
    throw new Error('Failed to load app info');
  }
  return body;
}

export default function _index() {
  const infoBody = useLoaderData();

  // useEffect(() => {
  //   //  fractiesClient.getAll().then(({status, body})=> {
  //   //     if (status === 200){
  //   //         console.log(body)
  //   //     }
  //   // })
  // }, [])


  return <>
      <Welcome appInfo={infoBody} />
      <MouseMoveEffect />
    </>
}
