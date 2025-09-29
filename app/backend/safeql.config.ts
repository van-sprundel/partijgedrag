import "dotenv/config";
import { defineConfig } from "@ts-safeql/eslint-plugin";

const databaseUrl = process.env.DATABASE_URL;

if (!databaseUrl) {
	throw new Error("DATABASE_URL environment variable is required");
}

export default defineConfig({
	connections: {
		databaseUrl,
		targets: [
			{ tag: "sql", transform: "{type}" },
			{ tag: "sqlOne", transform: "{type}" },
		],
		overrides: {
			types: {
				json: "unknown",
				jsonb: "unknown",
				uuid: "UUID",
			},
		},
	},
});
