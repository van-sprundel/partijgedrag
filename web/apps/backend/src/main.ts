import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module.js';
import dotenv from 'dotenv';
import { ValidationPipe } from '@nestjs/common';
import { GlobalExceptionFilter } from './global.response.exception_filter.js';
import { generateOpenApi } from '@ts-rest/open-api';
import { rootContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { SwaggerModule } from '@nestjs/swagger';
import cookieParser from 'cookie-parser';

dotenv.config();

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // app.enableCors({
  //   origin: '*',
  //   credentials: true,
  // });

  const document = generateOpenApi(rootContract, {
    info: {
      title: 'Full Stack App Template',
      version: '1.0.0',
      openapi: '3.0.0',
      components: {
        aaa: {
          type: 'apiKey',
          in: 'header',
          name: 'Authorization',
        },
      },
    },
    security: {
      aaa: [],
    },
  });

  SwaggerModule.setup('api-docs', app, document);
  app.use(cookieParser());
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true, // Strips properties that are not in the DTO
      forbidNonWhitelisted: true, // Throws error for unknown properties
      transform: true, // Automatically transforms payloads to DTO instances
    }),
  );

  app.useGlobalFilters(new GlobalExceptionFilter());

  await app.listen(process.env.BACKEND_PORT ?? 3000);
}
bootstrap();
