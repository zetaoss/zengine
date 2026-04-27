// eslint.config.js
import js from '@eslint/js'
import { defineConfig } from 'eslint/config'
import betterTailwindcss from 'eslint-plugin-better-tailwindcss'
import { getDefaultSelectors } from 'eslint-plugin-better-tailwindcss/api/defaults'
import { MatcherType, SelectorKind } from 'eslint-plugin-better-tailwindcss/api/types'
import simpleImportSort from 'eslint-plugin-simple-import-sort'
import svelte from 'eslint-plugin-svelte'
import globals from 'globals'
import ts from 'typescript-eslint'

import svelteConfig from './svelte.config.js'

const tailwindSelectors = [
  ...getDefaultSelectors(),
  {
    kind: SelectorKind.Variable,
    name: '.*Classes$',
    match: [{ type: MatcherType.String }, { type: MatcherType.ObjectValue }],
  },
]

export default defineConfig(
  {
    ignores: ['.svelte-kit/**', 'dist/**'],
  },
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig,
      },
    },
    rules: {
      'svelte/block-lang': ['error', { script: 'ts', style: [null, 'postcss'] }],
    },
  },
  {
    name: 'app/import-sorting',
    plugins: {
      'simple-import-sort': simpleImportSort,
    },
    rules: {
      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',
    },
  },
  {
    name: 'app/tailwind-shorthand',
    plugins: {
      'better-tailwindcss': betterTailwindcss,
    },
    rules: {
      'better-tailwindcss/enforce-consistent-variable-syntax': [
        'error',
        { entryPoint: './src/shared/assets/app.css', syntax: 'shorthand', selectors: tailwindSelectors },
      ],
    },
  },
)
