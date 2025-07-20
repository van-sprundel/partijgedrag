/** Types generated for queries found in "src/db/queries/moties.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

export type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

export type NumberOrString = number | string;

/** 'Get' parameters type */
export interface IGetParams {
  limit: NumberOrString;
  offset: NumberOrString;
}

/** 'Get' return type */
export interface IGetResult {
  actor_fractie: string | null;
  actor_naam: string | null;
  besluit_tekst: string;
  created_at: Date;
  gestart_op: Date | null;
  id: string;
  nummer: string;
  onderwerp: string;
  party_votes: Json;
  status: string;
  titel: string;
  updated_at: Date;
  vergaderjaar: string;
}

/** 'Get' query type */
export interface IGetQuery {
  params: IGetParams;
  result: IGetResult;
}

const getIR: any = {"usedParamSet":{"limit":true,"offset":true},"params":[{"name":"limit","required":true,"transform":{"type":"scalar"},"locs":[{"a":358,"b":364}]},{"name":"offset","required":true,"transform":{"type":"scalar"},"locs":[{"a":373,"b":380}]}],"statement":"SELECT z.id, z.nummer, z.onderwerp, z.titel, za.actor_fractie, za.actor_naam, z.updated_at, z.created_at, z.status, z.gestart_op, z.vergaderjaar, v.besluit_tekst, v.party_votes FROM \"zaken\" as z\nINNER JOIN \"zaak_actors\" as za ON z.id = za.zaak_id\nINNER JOIN \"besluiten\"  as b ON z.id = b.zaak_id\nINNER JOIN \"voting_results\" as v ON b.id = v.besluit_id\nLIMIT :limit! OFFSET :offset!"};

/**
 * Query generated from SQL:
 * ```
 * SELECT z.id, z.nummer, z.onderwerp, z.titel, za.actor_fractie, za.actor_naam, z.updated_at, z.created_at, z.status, z.gestart_op, z.vergaderjaar, v.besluit_tekst, v.party_votes FROM "zaken" as z
 * INNER JOIN "zaak_actors" as za ON z.id = za.zaak_id
 * INNER JOIN "besluiten"  as b ON z.id = b.zaak_id
 * INNER JOIN "voting_results" as v ON b.id = v.besluit_id
 * LIMIT :limit! OFFSET :offset!
 * ```
 */
export const get = new PreparedQuery<IGetParams,IGetResult>(getIR);


