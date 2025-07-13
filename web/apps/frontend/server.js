import { createRequestHandler } from "@react-router/express";
import compression from "compression";
import express from "express";
import {createProxyMiddleware} from "http-proxy-middleware";
import dotenv from "dotenv";

dotenv.config();

const viteDevServer =
    process.env.NODE_ENV === "production"
        ? undefined
        : await import("vite").then((vite) =>
            vite.createServer({
                server: { middlewareMode: true },
            })
        );


const remixHandler = createRequestHandler({
    build: viteDevServer
        ? () => viteDevServer.ssrLoadModule("virtual:react-router/server-build")
        : await import("./build/server/index.js"),
});

const app = express();

app.use(compression());




// http://expressjs.com/en/advanced/best-practice-security.html#at-a-minimum-disable-x-powered-by-header
app.disable("x-powered-by");


// handle asset requests
if (viteDevServer) {
    app.use(viteDevServer.middlewares);
} else {
    // Vite fingerprints its assets so we can cache forever.
    console.log('no viteDevServer')
    app.use(
        "/assets",
        express.static("build/client/assets", { immutable: true, maxAge: "1y" })
    );
}


const target = process.env.BACKEND_URL;
console.log('target', target)
const proxyMiddleware = createProxyMiddleware({
    target,
    changeOrigin: true,
});

app.use('/api', proxyMiddleware);

// Everything else (like favicon.ico) is cached for an hour. You may want to be
// more aggressive with this caching.
app.use(express.static("build/client", { maxAge: "1h" }));


// handle SSR requests
app.all("*", remixHandler);

const port = process.env.FRONTEND_PORT || 3000;
app.listen(port, '0.0.0.0', () =>{
    console.log(`Express server listening at http://localhost:${port}`)
});
