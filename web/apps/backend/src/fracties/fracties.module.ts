import { Module } from '@nestjs/common';
import { FractiesService } from './fracties.service.js';
import {FractiesController} from "./fracties.controller.js";

@Module({
  imports: [],
  providers: [FractiesService],
  exports: [FractiesService],
  controllers: [FractiesController]
})
export class FractiesModule {}
