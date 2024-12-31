/* eslint-env node */
require('@rushstack/eslint-patch/modern-module-resolution')

module.exports = {
  ignorePatterns: [
    'e2e/vue.spec.ts',
    'postcss.config.js',
    'tailwind.config.js',
    'vitest.config.ts',
  ],
  root: true,
  extends: [
    'plugin:vue/vue3-essential',
    '@vue/eslint-config-airbnb-with-typescript',
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    project: ['./tsconfig.json'],
  },
  rules: {
    '@typescript-eslint/consistent-type-imports': ['error', { prefer: 'type-imports' }],
    '@typescript-eslint/semi': 'off', // DEPRECATED https://typescript-eslint.io/rules/semi/
    'comma-dangle': ['error', 'always-multiline'],

    'import/no-extraneous-dependencies': 'off', // to use symlink
    'import/order': [
      'error', {
        pathGroups: [{ pattern: 'vue', group: 'builtin', position: 'before' }, { pattern: '@/**', group: 'parent', position: 'before' }],
        pathGroupsExcludedImportTypes: ['builtin'],
        alphabetize: { order: 'asc' },
        'newlines-between': 'always',
      },
    ],
    // 'import/order': 'off', // using sort-imports

    'no-alert': 'off', // someday
    'no-await-in-loop': 'off', // tradeoff(readability)
    'no-bitwise': 'off', // ok
    'no-console': 'off', // TODO
    'no-nested-ternary': 'off', // ok
    'no-restricted-syntax': ['error', 'LabeledStatement', 'WithStatement'], // tradeoff
    'no-param-reassign': 'off', // tradeoff
    'no-plusplus': 'off', // ok

    'padding-line-between-statements': 'error',
    semi: ['error', 'never'], // @typescript-eslint/semi
    'sort-imports': ['error', { ignoreDeclarationSort: true }],

    // https://eslint.vuejs.org/rules/
    // Priority B: Strongly Recommended
    'vue/first-attribute-linebreak': 'off', // volar
    'vue/html-closing-bracket-newline': 'off', // volar
    'vue/html-indent': 'off', // volar
    'vue/max-attributes-per-line': 'off', // volar
    'vue/multiline-html-element-content-newline': 'error',
    // Priority C: Recommended
    'vue/no-lone-template': 'error',
    'vue/no-v-html': 'off', // v-html
    // Extension Rules
    'vue/max-len': 'off', // volar

    // https://vue-a11y.github.io/eslint-plugin-vuejs-accessibility/
    'vuejs-accessibility/anchor-has-content': 'off', // false alarm
    'vuejs-accessibility/form-control-has-label': 'off', // not sure
    'vuejs-accessibility/no-access-key': 'off', // mediawiki
  },
}