# ESLint Flat Config Reference

## eslint.config.mjs — Basic Structure

```js
// @ts-check
import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import globals from 'globals'

export default tseslint.config(
  // Global ignores (replaces .eslintignore)
  {
    ignores: ['dist/**', 'build/**', '.next/**', 'coverage/**', '*.min.js'],
  },

  // Base JS recommended
  js.configs.recommended,

  // TypeScript — pick one preset:
  ...tseslint.configs.recommended,       // basic TS rules, no type info needed
  // ...tseslint.configs.strict,         // stricter superset of recommended
  // ...tseslint.configs.stylistic,      // opinionated style rules

  // Typed linting (requires parserOptions.projectService)
  // ...tseslint.configs.recommendedTypeChecked,
  // ...tseslint.configs.strictTypeChecked,
  // ...tseslint.configs.stylisticTypeChecked,

  // Global config for all files
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2022,
        ...globals.node,
      },
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    rules: {
      'no-console': 'warn',
      'no-unused-vars': 'off',           // use TS version instead
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'error',
    },
  },
)
```

## Typed Linting (parserOptions.projectService)

```js
// eslint.config.mjs
import tseslint from 'typescript-eslint'

export default tseslint.config(
  ...tseslint.configs.recommendedTypeChecked,

  {
    languageOptions: {
      parserOptions: {
        projectService: true,              // auto-discovers tsconfig.json
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // Rules that require type info (only available with typed linting)
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/await-thenable': 'error',
      '@typescript-eslint/no-unnecessary-type-assertion': 'error',
      '@typescript-eslint/prefer-nullish-coalescing': 'error',
      '@typescript-eslint/prefer-optional-chain': 'error',
    },
  },

  // Disable type-checked rules for JS files (they can't be type-checked)
  {
    files: ['**/*.js', '**/*.mjs', '**/*.cjs'],
    ...tseslint.configs.disableTypeChecked,
  },
)
```

## File-Specific Overrides

```js
export default tseslint.config(
  // Base config
  js.configs.recommended,
  ...tseslint.configs.recommended,

  // Override for test files
  {
    files: ['**/*.test.ts', '**/*.spec.ts', '**/*.test.tsx', 'src/test/**'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      'no-console': 'off',
    },
  },

  // Override for config files (Node.js globals)
  {
    files: ['*.config.ts', '*.config.js', '*.config.mjs'],
    languageOptions: {
      globals: globals.node,
    },
    rules: {
      '@typescript-eslint/no-var-requires': 'off',
    },
  },

  // Override for specific directories
  {
    files: ['scripts/**'],
    rules: {
      'no-console': 'off',
    },
  },
)
```

## Custom Plugin Integration

```js
import pluginReact from 'eslint-plugin-react'
import pluginReactHooks from 'eslint-plugin-react-hooks'
import pluginImport from 'eslint-plugin-import'

export default tseslint.config(
  // React plugin with flat config
  {
    plugins: {
      react: pluginReact,
      'react-hooks': pluginReactHooks,
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      ...pluginReact.configs.recommended.rules,
      ...pluginReactHooks.configs.recommended.rules,
      'react/react-in-jsx-scope': 'off',     // not needed with React 17+
      'react/prop-types': 'off',             // TS handles this
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
    },
  },

  // Import plugin
  {
    plugins: { import: pluginImport },
    rules: {
      'import/order': ['error', {
        groups: ['builtin', 'external', 'internal', 'parent', 'sibling', 'index'],
        'newlines-between': 'always',
      }],
      'import/no-duplicates': 'error',
    },
  },
)
```

## Config Object Shape

```ts
// Each element in the config array is a ConfigObject:
type ConfigObject = {
  name?: string                   // label for debugging (shows in --debug output)
  files?: string[]                // glob patterns this config applies to
  ignores?: string[]              // when top-level: global ignores; otherwise: exclude within files
  plugins?: Record<string, Plugin>
  rules?: Record<string, RuleConfig>
  settings?: Record<string, unknown>
  processor?: string | Processor
  languageOptions?: {
    parser?: Parser
    parserOptions?: ParserOptions
    globals?: Record<string, 'readonly' | 'writable' | 'off'>
    sourceType?: 'module' | 'commonjs' | 'script'
    ecmaVersion?: number | 'latest'
  }
  linterOptions?: {
    reportUnusedDisableDirectives?: boolean | 'error' | 'warn' | 'off'
    noInlineConfig?: boolean
  }
}
```

## Next.js + TypeScript Full Example

```js
// eslint.config.mjs
// @ts-check
import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import pluginNext from '@next/eslint-plugin-next'
import pluginReact from 'eslint-plugin-react'
import pluginReactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'

export default tseslint.config(
  { ignores: ['.next/**', 'out/**', 'coverage/**'] },

  js.configs.recommended,
  ...tseslint.configs.recommended,

  {
    plugins: {
      '@next/next': pluginNext,
      react: pluginReact,
      'react-hooks': pluginReactHooks,
    },
    languageOptions: {
      globals: { ...globals.browser, ...globals.es2022 },
    },
    settings: { react: { version: 'detect' } },
    rules: {
      ...pluginNext.configs.recommended.rules,
      ...pluginNext.configs['core-web-vitals'].rules,
      'react/react-in-jsx-scope': 'off',
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'error',
    },
  },

  // Relax rules for test files
  {
    files: ['**/*.test.{ts,tsx}', '**/*.spec.{ts,tsx}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
)
```

## CLI Usage

```bash
npx eslint .                        # lint all files
npx eslint src/ --ext .ts,.tsx      # specific dir (legacy, not needed in flat config)
npx eslint --fix .                  # auto-fix
npx eslint --max-warnings=0 .       # treat warnings as errors in CI
npx eslint --debug .                # show config resolution for each file
npx eslint --print-config src/app.ts  # show resolved config for a file
```
