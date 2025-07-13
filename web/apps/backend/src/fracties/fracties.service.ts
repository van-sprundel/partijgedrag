import { HttpException, HttpStatus, Injectable } from '@nestjs/common';
import { RegisterBody } from '@fullstack-typescript-template/ts-rest-contracts';
import { DbService } from '../db/db.service.js';

@Injectable()
export class FractiesService {
  constructor(private db: DbService) {}

  async getAll() {
    return await this.db.fracties.getFracties()
  }
}
