import {Card} from "~/components/ui/card";
import { initClient } from '@ts-rest/core';
import { authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { useLoaderData } from 'react-router';



export default function Profile({user}: {user: any}) {

  return <Card>
    <h4>Email: {user?.email}</h4>
  </Card>;
}