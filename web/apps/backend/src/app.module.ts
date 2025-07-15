import { Module } from '@nestjs/common';
import { AppController } from './app.controller.js';
import { AppService } from './app.service.js';
import { ConfigModule } from '@nestjs/config';
import { SendgridModule } from './sendgrid/sendgrid.module.js';
import { AuthController } from './auth/auth.controller.js';
import { AuthModule } from './auth/auth.module.js';
import { ScheduleModule } from '@nestjs/schedule';
import { DbModule } from './db/db.module.js';
import { FractiesController } from './fracties/fracties.controller.js';
import { FractiesModule } from './fracties/fracties.module.js';
import { MotiesController } from './moties/moties.controller.js';
import { MotiesModule } from './moties/moties.module.js';

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    SendgridModule,
    AuthModule,
    ScheduleModule.forRoot(),
    DbModule,
    FractiesModule,
    MotiesModule,
  ],
  controllers: [AppController, AuthController, FractiesController, MotiesController],
  providers: [AppService],
})
export class AppModule {}
