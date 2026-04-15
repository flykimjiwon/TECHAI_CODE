# Turborepo Reference

## turbo.json — Task Pipeline

```jsonc
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": [
    ".env",              // changes to these files invalidate ALL task caches
    "tsconfig.json"
  ],
  "globalEnv": [
    "NODE_ENV",          // env vars that affect all tasks (not in .env files)
    "CI"
  ],
  "tasks": {
    "build": {
      "dependsOn": ["^build"],       // ^ = must build dependencies first (topological)
      "inputs": ["src/**", "package.json", "tsconfig.json"],
      "outputs": [".next/**", "dist/**", "!.next/cache/**"],
      "env": ["NEXT_PUBLIC_*", "DATABASE_URL"]   // wildcards supported
    },

    "test": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "test/**", "vitest.config.ts"],
      "outputs": ["coverage/**"],
      "env": ["CI"]
    },

    "lint": {
      "dependsOn": [],               // no dependencies — run immediately in parallel
      "inputs": ["src/**", ".eslintrc*", "eslint.config.*"]
    },

    "dev": {
      "dependsOn": ["^build"],
      "cache": false,                // never cache — always run fresh
      "persistent": true             // long-running process (dev server)
    },

    "type-check": {
      "dependsOn": ["^type-check"],
      "inputs": ["src/**/*.ts", "src/**/*.tsx", "tsconfig.json"]
    },

    // Same-package dependency (no ^)
    "test:e2e": {
      "dependsOn": ["build"],        // run THIS package's build first, not deps
      "cache": false,
      "persistent": false
    },

    // Cross-package dependency (specific package#task)
    "deploy": {
      "dependsOn": ["api#build", "web#build"],
      "cache": false
    }
  }
}
```

## dependsOn Patterns

```jsonc
"dependsOn": ["^build"]       // topological: all workspace deps must finish build
"dependsOn": ["build"]        // same-package: this package's build must finish first
"dependsOn": ["api#build"]    // cross-package: specific pkg#task must finish
"dependsOn": []               // no deps: run in parallel with everything
```

## inputs & outputs

```jsonc
"inputs": [
  "src/**",                   // all files under src/
  "package.json",
  "tsconfig.json",
  "!src/**/*.test.ts"         // exclude test files from cache key
],
"outputs": [
  "dist/**",                  // cache these directories
  ".next/**",
  "!.next/cache/**"           // exclude from cache (too large / irrelevant)
]
// If inputs not set: all non-gitignored files in package
// If outputs not set: nothing cached (task still runs, just not stored)
```

## Filtering (--filter)

```bash
# Single package
turbo build --filter=@myapp/web
turbo build --filter=web          # short name also works

# Package + its dependencies
turbo build --filter=web...       # web and everything web depends on

# Package + its dependents
turbo build --filter=...web       # web and everything that depends on web

# Changed since branch
turbo build --filter="[main]"     # packages changed vs main branch
turbo build --filter="[HEAD^1]"   # changed in last commit

# Exclude
turbo build --filter=!docs        # all except docs
turbo build --filter=!./apps/legacy

# Combine
turbo build --filter=web --filter=api
turbo build --filter=web... --filter="[main]"

# Directory glob
turbo build --filter=./apps/*
turbo build --filter=./packages/*
```

## env Wildcards

```jsonc
"tasks": {
  "build": {
    "env": [
      "NEXT_PUBLIC_*",          // all vars starting with NEXT_PUBLIC_
      "DATABASE_URL",
      "!SECRET_*"               // explicitly exclude (negate)
    ]
  }
}
```

## Per-Package Override (extends)

```jsonc
// apps/web/turbo.json — override specific tasks for this package
{
  "extends": ["//"],             // // = root turbo.json
  "tasks": {
    "build": {
      "outputs": [".next/**", "!.next/cache/**"],
      "env": ["NEXT_PUBLIC_API_URL", "VERCEL_URL"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    }
  }
}
```

## Remote Caching

```bash
# Link to Vercel Remote Cache
npx turbo login
npx turbo link

# Use in CI
TURBO_TOKEN=<token> TURBO_TEAM=<team> turbo build

# Self-hosted (e.g., Turborepo Remote Cache server)
# turbo.json
{
  "remoteCache": {
    "apiUrl": "https://my-cache-server.com"
  }
}
```

## Common CLI Flags

```bash
turbo run build                   # run build task
turbo build                       # shorthand
turbo build --dry-run             # show what would run without running
turbo build --dry-run=json        # machine-readable dry run output
turbo build --graph               # visualize task graph (opens browser)
turbo build --graph=graph.png     # export graph image
turbo build --concurrency=4       # limit parallel tasks (default: 10)
turbo build --force               # bypass cache, always re-run
turbo build --no-cache            # run but don't write to cache
turbo build --output-logs=none    # suppress task output (just show summary)
turbo build --output-logs=new-only  # only show output for cache misses
turbo build --summarize           # write run summary JSON to .turbo/runs/
```

## Monorepo package.json Scripts

```jsonc
// root package.json
{
  "scripts": {
    "build": "turbo run build",
    "dev": "turbo run dev",
    "test": "turbo run test",
    "lint": "turbo run lint",
    "type-check": "turbo run type-check",
    "clean": "turbo run clean && rm -rf node_modules"
  },
  "devDependencies": {
    "turbo": "^2.0.0"
  }
}
```

## Workspace Structure

```
my-monorepo/
├── turbo.json
├── package.json
├── apps/
│   ├── web/
│   │   ├── turbo.json      # optional per-package override
│   │   └── package.json
│   └── api/
│       └── package.json
└── packages/
    ├── ui/
    │   └── package.json    # "name": "@myapp/ui"
    └── config/
        └── package.json    # "name": "@myapp/config"
```

```jsonc
// root package.json workspaces
{
  "workspaces": ["apps/*", "packages/*"]
}
```
