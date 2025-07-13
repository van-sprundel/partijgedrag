import { z } from 'zod';

export function Secured() {
  return {
    headers: z.object({
      authorization: z.string(),
    }),
  };
}
