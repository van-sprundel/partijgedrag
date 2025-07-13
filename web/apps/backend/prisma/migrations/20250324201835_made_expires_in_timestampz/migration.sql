/*
  Warnings:

  - You are about to drop the column `expires` on the `UserSession` table. All the data in the column will be lost.
  - Added the required column `expiresAt` to the `UserSession` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "UserSession" DROP COLUMN "expires",
ADD COLUMN     "expiresAt" TIMESTAMPTZ(3) NOT NULL;
