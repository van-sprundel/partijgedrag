/** Types generated for queries found in "src/db/queries/fracties.sql" */
import { PreparedQuery } from '@pgtyped/runtime';

/** 'GetFracties' parameters type */
export type IGetFractiesParams = void;

/** 'GetFracties' return type */
export interface IGetFractiesResult {
  aantal_stemmen: number | null;
  aantal_zetels: number | null;
  afkorting: string | null;
  api_gewijzigd_op: Date | null;
  content_length: number | null;
  content_type: string | null;
  created_at: Date;
  datum_actief: Date | null;
  datum_inactief: Date | null;
  gewijzigd_op: Date | null;
  id: string;
  naam_en: string | null;
  naam_nl: string | null;
  nummer: number | null;
  updated_at: Date;
  verwijderd: boolean;
}

/** 'GetFracties' query type */
export interface IGetFractiesQuery {
  params: IGetFractiesParams;
  result: IGetFractiesResult;
}

const getFractiesIR: any = {"usedParamSet":{},"params":[],"statement":"SELECT * FROM fracties"};

/**
 * Query generated from SQL:
 * ```
 * SELECT * FROM fracties
 * ```
 */
export const getFracties = new PreparedQuery<IGetFractiesParams,IGetFractiesResult>(getFractiesIR);


