/* @name getSessionById */
SELECT * FROM user_sessions WHERE "id" = :id;

/* @name getSessionByAccessToken */
SELECT * FROM user_sessions WHERE "accessToken" = :accessToken;

/*  @name getSessionByRefreshToken */
SELECT * FROM user_sessions WHERE "refreshToken" = :refreshToken;

/* @name createSession */
INSERT INTO user_sessions ("userId", "expiresAt", "accessToken", "refreshToken")
VALUES (:userId!, :expiresAt!, :accessToken!, :refreshToken!)
RETURNING *;

/* @name deleteSessionByAccessToken */
DELETE FROM user_sessions WHERE "accessToken" = :accessToken!;

/* @name deleteAllSessionsOfUser */
DELETE FROM user_sessions WHERE "userId" = :userId!;

/* @name deleteExpiredSessions */
DELETE FROM user_sessions WHERE "expiresAt" < now();