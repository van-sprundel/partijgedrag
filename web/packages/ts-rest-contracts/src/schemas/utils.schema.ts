import { z, ZodRawShape } from 'zod';

export const pageSchema = z.object({
  pageSize: z.coerce.number().min(1).max(100).optional(),
  page: z.coerce.number().min(0).optional(),
});

export function pageView<
  U ,
   T=ZodRawShape ,
>(
  c: { type: <V>() => any },  // Adjust this to your actual contract helper type
  path: string = '/',
  query?: T
) {
  return {
    method: 'GET' as const,
    path,
    query: query ? pageSchema.extend(query) : pageSchema,
    responses: {
      200: c.type<{ data: U }>(),
    },
  };
}
