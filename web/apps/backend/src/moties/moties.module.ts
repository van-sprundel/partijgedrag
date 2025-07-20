import { Module } from '@nestjs/common';
import { MotiesService } from './moties.service.js';
import { MotiesController } from './moties.controller.js';

@Module({
  imports: [],
  providers: [MotiesService],
  exports: [MotiesService],
  controllers: [MotiesController],
})
export class MotiesModule {}
