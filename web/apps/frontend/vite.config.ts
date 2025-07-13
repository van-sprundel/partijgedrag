import { reactRouter } from '@react-router/dev/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  server: {
    port: 4000,
    allowedHosts: [
        'tunnel.merlijnmac.nl'
    ]
  },
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],
});
