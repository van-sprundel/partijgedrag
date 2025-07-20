/*
  Warnings:

  - Added the required column `token` to the `PasswordResetRequest` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "PasswordResetRequest" ADD COLUMN     "token" TEXT NOT NULL;
