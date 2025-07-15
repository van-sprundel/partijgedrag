import { DbService } from '../db/db.service.js';
import { Injectable } from '@nestjs/common';

@Injectable()
export class MotiesService {
  constructor(private db: DbService) {}

  async getAll(page?: number, pageSize?: number) {
    const limit = pageSize ?? 10;
    const offset = (page ?? 0) * limit;

    return (
      await this.db.moties.get({
        limit,
        offset,
      })
    ).map((m) => ({
      id: m.id,
      onderwerp: m.onderwerp,
      titel: m.titel,
      status: m.status,
      gestartOp: m.gestart_op,
      vergaderjaar: m.vergaderjaar,
      createdAt: m.created_at,
      updatedAt: m.updated_at,
    }));
  }
}
