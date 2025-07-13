import { Global, Module } from '@nestjs/common';
import { DbService } from './db.service.js';

@Global()
@Module({
  providers: [DbService],
  exports: [DbService],
})
export class DbModule {}
