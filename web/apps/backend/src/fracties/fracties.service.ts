import { Injectable } from '@nestjs/common';
import { DbService } from '../db/db.service.js';

@Injectable()
export class FractiesService {
  constructor(private db: DbService) {}

  async getAll(page?: number, pageSize?: number) {
    const limit = pageSize ?? 10;
    const offset = (page ?? 0) * limit;

    return await this.db.fracties.get({
      limit,
      offset,
    });
  }
}
