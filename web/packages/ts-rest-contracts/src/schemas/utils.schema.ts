import { z } from 'zod';

export const pageSchema = z.object({
  pageSize: z.coerce.number().min(1).max(100).optional(),
  page: z.coerce.number().min(0).optional(),
});