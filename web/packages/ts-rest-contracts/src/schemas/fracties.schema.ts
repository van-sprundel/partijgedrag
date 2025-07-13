import { z } from 'zod';

// ISO date-time strings are typical in JSON; if using Date objects directly, adjust accordingly
export const zFractieSchema = z.object({
  id: z.string(),
  nummer: z.number().int().nullable().optional(),
  afkorting: z.string().nullable().optional(),
  naam_nl: z.string().nullable().optional(),
  naam_en: z.string().nullable().optional(),
  aantal_zetels: z.number().int().nullable().optional(),
  aantal_stemmen: z.number().int().nullable().optional(),
  datum_actief: z.coerce.date().nullable().optional(),
  datum_inactief: z.coerce.date().nullable().optional(),
  content_type: z.string().nullable().optional(),
  content_length: z.number().int().nullable().optional(),
  gewijzigd_op: z.coerce.date().nullable().optional(),
  api_gewijzigd_op: z.coerce.date().nullable().optional(),
  verwijderd: z.boolean(),
  created_at: z.coerce.date(),
  updated_at: z.coerce.date(),
});

export type Fractie = z.infer<typeof zFractieSchema>