import { Module } from '@nestjs/common';
import { SendgridService } from './sendgrid.service.js';
import { ConfigModule } from '@nestjs/config';

@Module({
  imports: [ConfigModule],
  providers: [SendgridService],
  exports: [SendgridService],
})
export class SendgridModule {}
