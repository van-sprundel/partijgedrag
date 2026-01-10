import "dotenv/config";
import safeql from "@ts-safeql/eslint-plugin/config";
import tseslint from "typescript-eslint";

const databaseUrl = process.env.DATABASE_URL;

if (!databaseUrl) {
  throw new Error("DATABASE_URL environment variable is required");
}

export default tseslint.config(
  {
    ignores: ['dist/**', 'node_modules/**', '**/*.js', '**/*.mjs', '**/*.cjs']
  },
  ...tseslint.configs.recommended,
  safeql.configs.connections({
    databaseUrl,
    targets: [
      { tag: "sql", transform: "{type}" },
      { tag: "sqlOne", transform: "{type}" },
      { tag: "sqlOneOrNull", transform: "{type}" },
    ],
    overrides: {
      types: {
        json: "unknown",
        jsonb: "unknown",
        uuid: "`${string}-${string}-${string}-${string}-${string}`",
        bytea: "unknown",
      },
    },
  }),
  {
    rules: {
      "@ts-safeql/check-sql": "off",
    },
  },
);
