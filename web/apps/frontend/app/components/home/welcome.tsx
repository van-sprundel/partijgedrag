import { initClient } from '@ts-rest/core';
import MotionCard from '~/components/motion';
import { motieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/moties.contract';
import { useEffect, useState } from 'react';
import type { Motie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/moties.schema';
import { Container, Flex } from '@radix-ui/themes';




export const API_BASE_URL = 'http://localhost:3000/api'

export const client = initClient(motieContract, {
    baseUrl: '/api',
    baseHeaders: {
        'x-app-source': 'ts-rest',
    },
});




export function Welcome()
{
  const [moties, setMoties] = useState<Motie[]>([]);

  useEffect(() => {
    client.getAll().then((res) => {
        if(res.status === 200) {
            setMoties(res.body.moties);
        }
    });
  },[])

  return (
      <Container>
        <Flex gap="3" align="center" direction='column' width={'100%'}>
            {
                moties.length > 0 && moties.map((motie) => (
                    <MotionCard motie={motie} />
                ))
            }
            </Flex>
      </Container>
  );
}


