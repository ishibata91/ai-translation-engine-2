import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import importPlugin from 'eslint-plugin-import';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import prettierConfig from 'eslint-config-prettier';

export default tseslint.config(
  {
    ignores: ['dist/**', 'wailsjs/**', 'node_modules/**'],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['src/**/*.{ts,tsx}'],
    plugins: {
      import: importPlugin,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'no-useless-escape': 'off',
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      'import/no-restricted-paths': [
        'error',
        {
          zones: [
            {
              target: './src/pages',
              from: './wailsjs',
              message: 'pages から wailsjs への直接 import を禁止。hooks/features 経由にしてください。',
            },
            {
              target: './src/pages',
              from: './src/store',
              message: 'pages から store への直接 import を禁止。hooks/features 経由にしてください。',
            },
          ],
        },
      ],
    },
  },
  {
    files: [
      'src/pages/DictionaryBuilder.tsx',
      'src/hooks/features/dictionaryBuilder/**/*.{ts,tsx}',
    ],
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      'react-hooks/exhaustive-deps': 'error',
      'max-depth': ['error', 4],
      'no-else-return': ['error', { allowElseIf: false }],
      'no-console': ['warn', { allow: ['warn', 'error'] }],
    },
  },
  prettierConfig,
);
