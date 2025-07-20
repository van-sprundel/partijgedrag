/* @name getPasswordResetRequestByUserId */
SELECT * FROM password_reset_requests WHERE "userId" = :userId;

/* @name createPasswordResetRequest */
INSERT INTO password_reset_requests ("userId", "token", "expiresAt") VALUES (:userId!, :token!, :expiresAt!)
RETURNING *;

/* @name deletePasswordResetRequestByUserId */
DELETE FROM password_reset_requests WHERE "userId" = :userId!;

/* @name deleteExpiredPasswordResetRequests */
DELETE FROM password_reset_requests WHERE "expiresAt" < now();

/* @name getPasswordResetRequestByToken */
SELECT * FROM password_reset_requests WHERE "token" = :token;

/* @name deletePasswordResetRequestByToken */
DELETE FROM password_reset_requests WHERE "token" = :token!;
