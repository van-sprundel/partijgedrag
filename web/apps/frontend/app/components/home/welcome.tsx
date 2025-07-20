import { initClient } from '@ts-rest/core';
import MotieCard from '~/components/moties/motie-card';
import { motieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/moties.contract';
import { useEffect, useState } from 'react';
import type { Motie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/moties.schema';
import { Button, Container, DropdownMenu, Flex, IconButton, ScrollArea } from '@radix-ui/themes';
import { useSearchParams } from 'react-router';
import { ArrowLeft } from 'lucide-react';
import { ArrowLeftIcon, ArrowRightIcon } from '@radix-ui/react-icons';
import { CardView } from '~/components/card-view';




export const API_BASE_URL = 'http://localhost:3000/api'

export const client = initClient(motieContract, {
    baseUrl: '/api',
    baseHeaders: {
        'x-app-source': 'ts-rest',
    },
});



const defaultSearchParams = {
    page: "0",
    pageSize: "5"
}

type SearchParams = {
    page: string,
    pageSize: string
}

function transformPageQuery(params: URLSearchParams) {
    const page = Number(params.get('page')) ?? 0;
    const pageSize = Number(params.get('pageSize')) ?? 5;


    return {
        page,
        pageSize
    }
}

export function Welcome()
{
    return (
        <CardView<Motie>
            queryFn={(search) => client.getAll({
                query: {
                    ...transformPageQuery(search),
                }
            }).then((res) => {
                if(res.status === 200) {
                    return res.body.moties;
                }
            })}
            renderCard={(motie) => (
                <MotieCard motie={motie} />
            )}
        />
    );

  // const [search, setSearch] = useSearchParams(defaultSearchParams);
  // const [moties, setMoties] = useState<Motie[]>([]);
  //
  // useEffect(() => {
  //   client.getAll({
  //       query: {
  //           ...transformPageQuery(search),
  //       }
  //   }).then((res) => {
  //       if(res.status === 200) {
  //           setMoties(res.body.moties);
  //       }
  //   });
  // },[search])
  //
  //   function prevPage() {
  //       setSearch({
  //           page: (Number(search.get('page')) - 1).toString(),
  //           pageSize: (Number(search.get('pageSize'))).toString()
  //       })
  //   }
  //
  //   function nextPage() {
  //       setSearch({
  //           page: (Number(search.get('page')) + 1).toString(),
  //           pageSize: (Number(search.get('pageSize'))).toString()
  //       })
  //   }
  //
  //   function setPageSize(pageSize: number) {
  //       setSearch({
  //           page: (Number(search.get('page'))).toString(),
  //           pageSize: pageSize.toString()
  //       })
  //   }
  //
  // return (
  //     <Container>
  //         <Flex gap="3" align="center" justify={'center'} direction='column' width={'100%'}>
  //         <Flex gap="3" align="center" justify={'center'} direction='row' width={'100%'}>
  //             <IconButton onClick={prevPage}>
  //                 <ArrowLeftIcon width="18" height="18" />
  //             </IconButton>
  //             <DropdownMenu.Root >
  //                 <DropdownMenu.Trigger>
  //                     <Button variant="soft" >
  //                         {search.get('pageSize')}
  //                         <DropdownMenu.TriggerIcon />
  //                     </Button>
  //                 </DropdownMenu.Trigger>
  //                 <DropdownMenu.Content >
  //                     {[5, 10, 20, 50, 100].map((size) => (
  //                         <DropdownMenu.Item
  //                             key={size}
  //                             onSelect={() => setPageSize(size)}
  //                         >
  //                             {size}
  //                         </DropdownMenu.Item>
  //                     ))}
  //                 </DropdownMenu.Content>
  //             </DropdownMenu.Root>
  //             <IconButton onClick={nextPage}>
  //                 <ArrowRightIcon width="18" height="18" />
  //             </IconButton>
  //         </Flex>
  //         <ScrollArea type="always" scrollbars="vertical" style={{ height: 700 }}>
  //
  //       <Flex gap="3" align="center" direction='column' width={'100%'}>
  //           {
  //               moties.length > 0 && moties.map((motie) => (
  //                   <MotionCard motie={motie} />
  //               ))
  //           }
  //           </Flex>
  //         </ScrollArea>
  //         </Flex>
  //     </Container>
  // );
}


