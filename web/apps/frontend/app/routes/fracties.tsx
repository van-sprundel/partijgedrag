
import type { Motie } from '@fullstack-typescript-template/ts-rest-contracts/dist/schemas/moties.schema';
import { CardView } from '~/components/card-view';
import { initClient } from '@ts-rest/core';
import { fractieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/fracties.contract';
import { FractieCard } from '~/components/fracties/fractie-card';





export const client = initClient(fractieContract, {
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

export default function Fracties() {
    return (
        <CardView<Motie>
            queryFn={(search) => client.getAll({
                query: {
                    ...transformPageQuery(search),
                }
            }).then((res) => {
                if (res.status === 200) {
                    return res.body.moties;
                }
            })}
            renderCard={(fractie) => (
                <FractieCard fractie={fractie} />
            )}
        />
    );
}