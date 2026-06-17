module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react-hooks/recommended',
  ],
  ignorePatterns: ['dist', '.eslintrc.cjs'],
  parser: '@typescript-eslint/parser',
  plugins: ['react-refresh'],
  rules: {
    // Fast Refresh の対応チェック
    'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    // TypeScript の不要な any を禁止
    '@typescript-eslint/no-explicit-any': 'error',
    // 未使用変数を禁止
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
  },
};
