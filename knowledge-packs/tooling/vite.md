# Vite Config Reference

## vite.config.ts Skeleton

```ts
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig(({ command, mode }) => {
  // Load env file based on mode ('development' | 'production' | 'test')
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [
      react(),
      // other plugins...
    ],

    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@components': path.resolve(__dirname, './src/components'),
        '~': path.resolve(__dirname, './'),
      },
    },

    define: {
      // Replace at build time (no quotes = raw JS; add JSON.stringify for strings)
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      'process.env.NODE_ENV': JSON.stringify(mode),
    },

    server: {
      port: 3000,
      strictPort: true,          // fail if port taken (don't auto-increment)
      open: true,                // open browser on start
      host: '0.0.0.0',          // expose to network
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
          secure: false,
        },
        '/ws': {
          target: 'ws://localhost:8080',
          ws: true,
        },
      },
      hmr: {
        overlay: true,           // show error overlay
        port: 24678,             // separate HMR port (useful behind proxies)
      },
      watch: {
        ignored: ['**/node_modules/**', '**/.git/**'],
      },
    },

    build: {
      target: 'es2020',          // 'esnext' | 'es2015' | ['chrome87', 'firefox78']
      outDir: 'dist',
      assetsDir: 'assets',
      sourcemap: true,           // true | false | 'inline' | 'hidden'
      minify: 'esbuild',         // 'esbuild' | 'terser' | false
      cssMinify: true,

      rollupOptions: {
        input: {
          main: path.resolve(__dirname, 'index.html'),
          // multi-page
          admin: path.resolve(__dirname, 'admin.html'),
        },
        output: {
          manualChunks: {
            vendor: ['react', 'react-dom'],
            router: ['react-router-dom'],
          },
          // or as function:
          // manualChunks(id) {
          //   if (id.includes('node_modules')) return 'vendor'
          // }
          chunkFileNames: 'assets/[name]-[hash].js',
          assetFileNames: 'assets/[name]-[hash][extname]',
        },
      },

      // Library mode
      lib: {
        entry: path.resolve(__dirname, 'src/index.ts'),
        name: 'MyLib',
        fileName: (format) => `my-lib.${format}.js`,
        formats: ['es', 'cjs', 'umd'],
      },

      chunkSizeWarningLimit: 500,  // kB
    },

    preview: {
      port: 4173,
      strictPort: true,
    },

    optimizeDeps: {
      include: ['lodash-es'],      // force pre-bundle
      exclude: ['@vite/client'],
    },

    css: {
      modules: {
        localsConvention: 'camelCaseOnly',
      },
      preprocessorOptions: {
        scss: {
          additionalData: `@import "@/styles/variables.scss";`,
        },
      },
    },
  }
})
```

## Environment Variables

```
# .env                  — loaded in all cases
# .env.local            — loaded in all cases, git-ignored
# .env.[mode]           — only loaded in specified mode
# .env.[mode].local     — only in specified mode, git-ignored
# Priority: .env.[mode].local > .env.[mode] > .env.local > .env
```

```bash
# .env
VITE_API_URL=https://api.example.com
VITE_APP_TITLE=My App
SECRET_KEY=not-exposed               # No VITE_ prefix = server-only, NOT in bundle
```

```ts
// In client code — only VITE_ prefixed vars are exposed
const apiUrl = import.meta.env.VITE_API_URL
const isDev  = import.meta.env.DEV          // boolean
const isProd = import.meta.env.PROD         // boolean
const mode   = import.meta.env.MODE         // 'development' | 'production' | custom
const base   = import.meta.env.BASE_URL     // from vite config base option

// TypeScript IntelliSense — vite-env.d.ts
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_APP_TITLE: string
  // add all your VITE_ vars here
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
```

## loadEnv for Server-Side / Config Files

```ts
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  // Load ALL env vars (prefix '' = no filter, includes non-VITE_ vars)
  const env = loadEnv(mode, process.cwd(), '')

  return {
    define: {
      __API_URL__: JSON.stringify(env.API_URL),  // can use non-VITE_ vars here
    },
    plugins: [
      myPlugin({ apiKey: env.SECRET_API_KEY }),  // server-side usage in plugin
    ],
  }
})
```

## Common Plugins

```ts
import react from '@vitejs/plugin-react'                  // React + HMR
import { TanStackRouterVite } from '@tanstack/router-vite-plugin'
import tsconfigPaths from 'vite-tsconfig-paths'           // use tsconfig paths
import { visualizer } from 'rollup-plugin-visualizer'     // bundle analysis
import { VitePWA } from 'vite-plugin-pwa'                 // PWA support
import checker from 'vite-plugin-checker'                 // TS/ESLint in dev

export default defineConfig({
  plugins: [
    react({
      babel: { plugins: ['babel-plugin-react-compiler'] }, // React 19 compiler
    }),
    tsconfigPaths(),
    checker({ typescript: true }),
    visualizer({ open: true }),  // generates stats.html after build
  ],
})
```

## Test Mode (Vitest)

```ts
// vitest.config.ts — extends vite config
import { defineConfig, mergeConfig } from 'vitest/config'
import viteConfig from './vite.config'

export default mergeConfig(viteConfig, defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
}))
```
