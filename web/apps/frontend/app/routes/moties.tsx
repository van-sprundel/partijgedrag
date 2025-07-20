import { initClient } from '@ts-rest/core';
import MotieCard from '~/components/moties/motie-card';
import { motieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/moties.contract';
import type { Motie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/moties.schema';
import { CardView } from '~/components/card-view';

export const client = initClient(motieContract, {
    baseUrl: '/api',
    baseHeaders: {
        'x-app-source': 'ts-rest',
    },
});


function transformPageQuery(params: URLSearchParams) {
    const page = Number(params.get('page')) ?? 0;
    const pageSize = Number(params.get('pageSize')) ?? 5;

    return {
        page,
        pageSize
    }
}

export default function Moties()
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
}


