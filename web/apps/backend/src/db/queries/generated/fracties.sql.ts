/** Types generated for queries found in "src/db/queries/fracties.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

export type NumberOrString = number | string;

/** 'Get' parameters type */
export interface IGetParams {
  limit: NumberOrString;
  offset: NumberOrString;
}

/** 'Get' return type */
export interface IGetResult {
  aantal_stemmen: string | null;
  aantal_zetels: string | null;
  afkorting: string | null;
  api_gewijzigd_op: Date | null;
  content_length: string | null;
  content_type: string | null;
  created_at: Date | null;
  datum_actief: Date | null;
  datum_inactief: Date | null;
  gewijzigd_op: Date | null;
  id: string;
  naam_en: string | null;
  naam_nl: string | null;
  nummer: string | null;
  updated_at: Date;
  verwijderd: boolean;
}

/** 'Get' query type */
export interface IGetQuery {
  params: IGetParams;
  result: IGetResult;
}

const getIR: any = {"usedParamSet":{"limit":true,"offset":true},"params":[{"name":"limit","required":true,"transform":{"type":"scalar"},"locs":[{"a":34,"b":40}]},{"name":"offset","required":true,"transform":{"type":"scalar"},"locs":[{"a":49,"b":56}]}],"statement":"SELECT * FROM fracties as z\nLIMIT :limit! OFFSET :offset!"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM fracties as z
 * LIMIT :limit! OFFSET :offset!
 * ```
 */
export const get = new PreparedQuery<IGetParams,IGetResult>(getIR);


