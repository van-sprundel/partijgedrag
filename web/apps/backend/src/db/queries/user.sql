/* @name getUnsafeUserByEmail */
SELECT * from users WHERE email = :email!;

/* @name getUserByEmail */
SELECT id, email from users WHERE email = :email!;

/* @name createUser */
INSERT INTO users ("email",  "password")
VALUES(:email!, :password!);

/* @name updatePassword */
UPDATE users
SET "password" = :password!
WHERE id = :id!;