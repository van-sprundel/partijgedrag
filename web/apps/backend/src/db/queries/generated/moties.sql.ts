/** Types generated for queries found in "src/db/queries/moties.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

export type NumberOrString = number | string;

/** 'Get' parameters type */
export interface IGetParams {
  limit: NumberOrString;
  offset: NumberOrString;
}

/** 'Get' return type */
export interface IGetResult {
  afgedaan: boolean;
  alias: string | null;
  api_gewijzigd_op: Date | null;
  citeertitel: string | null;
  created_at: Date | null;
  datum: Date | null;
  datum_afgedaan: Date | null;
  gestart_op: Date | null;
  gewijzigd_op: Date | null;
  grondslagvoorhang: string | null;
  groot_project: boolean;
  huidige_behandelstatus: string | null;
  id: string;
  kabinetsappreciatie: string;
  kamer: string;
  nummer: string;
  onderwerp: string;
  organisatie: string;
  soort: string;
  status: string;
  termijn: string | null;
  titel: string;
  updated_at: Date;
  vergaderjaar: string;
  verwijderd: boolean;
  volgnummer: string | null;
}

/** 'Get' query type */
export interface IGetQuery {
  params: IGetParams;
  result: IGetResult;
}

const getIR: any = {"usedParamSet":{"limit":true,"offset":true},"params":[{"name":"limit","required":true,"transform":{"type":"scalar"},"locs":[{"a":33,"b":39}]},{"name":"offset","required":true,"transform":{"type":"scalar"},"locs":[{"a":48,"b":55}]}],"statement":"SELECT * FROM \"zaken\" as z\nLIMIT :limit! OFFSET :offset!"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM "zaken" as z
 * LIMIT :limit! OFFSET :offset!
 * ```
 */
export const get = new PreparedQuery<IGetParams,IGetResult>(getIR);


