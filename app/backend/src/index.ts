import { os } from "@orpc/server";
import { RPCHandler } from "@orpc/server/node";
import cors from "cors";
import express from "express";
import { apiContract } from "./contracts/index.js";
import { db } from "./lib/db.js";
import { compassRouter } from "./routers/compass.js";
import { motionRouter } from "./routers/motions.js";
import { partyRouter } from "./routers/parties.js";

const app = express();
const port = process.env.PORT || 3001;

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
});

const handler = new RPCHandler(router);

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

// Create the main API router

// Health check endpoint
app.get("/health", (req, res) => {
	res.json({ status: "ok", timestamp: new Date().toISOString() });
});

// Basic route for testing
app.get("/", (req, res) => {
	res.json({
		message: "Partijgedrag API Server",
		version: "1.0.0",
		endpoints: {
			api: "/api",
			health: "/health",
		},
	});
});

// Error handling middleware
app.use(
	(
		err: Error,
		req: express.Request,
		res: express.Response,
		next: express.NextFunction,
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
app.listen(port, () => {
	console.log(`🚀 Server running on http://localhost:${port}`);
	console.log(`📡 API available at http://localhost:${port}/api`);
	console.log(`🔍 Health check at http://localhost:${port}/health`);
});

// Graceful shutdown
process.on("SIGINT", async () => {
	console.log("🔄 Gracefully shutting down...");
	await db.$disconnect();
	process.exit(0);
});

process.on("SIGTERM", async () => {
	console.log("🔄 Gracefully shutting down...");
	await db.$disconnect();
	process.exit(0);
});
