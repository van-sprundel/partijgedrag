import path from "node:path";
import { ORPCError, onError, os } from "@orpc/server";
import { RPCHandler } from "@orpc/server/node";
import cors from "cors";
import express from "express";
import { handleUriError } from "./middleware/handleUriError.js";
import { analysisRouter } from "./routers/analysis.js";
import { compassRouter } from "./routers/compass.js";
import { motionRouter } from "./routers/motions.js";
import { partyRouter } from "./routers/parties.js";
import { statisticsRouter } from "./routers/statistics.js";
import { pool } from "./services/db/client.js";

const app = express();
const host = process.env.HOST || "0.0.0.0";
const port = Number(process.env.PORT ?? 3001);

// CORS configuration
app.use(
	cors({
		origin: process.env.CORS_ORIGIN || "http://localhost:3000",
		credentials: true,
	}),
);

export const router = os.router({
	motions: motionRouter,
	parties: partyRouter,
	compass: compassRouter,
	statistics: statisticsRouter,
	analysis: analysisRouter,
});

const handler = new RPCHandler(router, {
	interceptors: [
		onError((error) => {
			if (error instanceof ORPCError) {
				console.error(JSON.stringify(error.cause));
				return;
			}
			console.error(error);
		}),
	],
});

app.use("/api*", async (req, res, next) => {
	const { matched } = await handler.handle(req, res, {
		prefix: "/api",
		context: {},
	});

	if (matched) {
		return;
	}

	next();
});

app.use(express.json());

// Serve frontend
const __dirname = path.dirname(new URL(import.meta.url).pathname);
app.use(express.static(path.join(__dirname, "../public")));

// Health check endpoint
app.get("/health", (_req, res) => {
	res.json({ status: "ok", timestamp: new Date().toISOString() });
});

app.get("/", (_req, res) => {
	res.json({
		message: "Partijgedrag API Server",
		version: "1.0.0",
		endpoints: {
			api: "/api",
			health: "/health",
		},
	});
});

// Catch-all to serve index.html
app.get("*", (_req, res) => {
	res.sendFile(path.join(__dirname, "../public", "index.html"));
});

// Error handling middleware
app.use(handleUriError);
app.use(
	(
		err: Error,
		_req: express.Request,
		res: express.Response,
		_next: express.NextFunction,
	) => {
		console.error("Server error:", err);
		res.status(500).json({
			error: "Internal server error",
			message:
				process.env.NODE_ENV === "development"
					? err.message
					: "Something went wrong",
		});
	},
);

// Start server
app.listen(port, host, () => {
	console.log(`🚀 Server running on http://0.0.0.0:${port}`);
	console.log(`📡 API available at http://0.0.0.0:${port}/api`);
	console.log(`🔍 Health check at http://0.0.0.0:${port}/health`);
});

const handleClose = async () => {
	await pool.end();
	process.exit(0);
};

process.on("SIGINT", handleClose);
process.on("SIGTERM", handleClose);
