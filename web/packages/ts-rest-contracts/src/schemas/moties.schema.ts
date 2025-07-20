import { z } from 'zod';

// ISO date-time strings are typical in JSON; if using Date objects directly, adjust accordingly
export const zMotieschema = z.object({
  id: z.string(),
  onderwerp: z.string(),
  titel: z.string(),
  status: z.string(),
  gestartOp: z.coerce.date().nullable(),
  vergaderjaar: z.string(),
  naamActor: z.string(),
  fractieActor: z.string(),
  besluitTekst: z.string(),
  partyVotes: z.array(z.string()),
  createdAt: z.coerce.date().nullable(),
  updatedAt: z.coerce.date(),
});

export type Motie = z.infer<typeof zMotieschema>