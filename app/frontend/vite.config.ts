import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const BACKEND_URL = "http://localhost:3001";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
	plugins: [react()],
	server: {
		port: 3000,
		proxy: {
			"/api": {
				target: BACKEND_URL,
				changeOrigin: true,
			},
		},
	},
	resolve: {
		alias: {
			"@": "/src",
		},
	},
	define: {
		"import.meta.env.VITE_API_URL":
			mode === "production" ? '""' : JSON.stringify(BACKEND_URL),
	},
}));
