import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import importPlugin from 'eslint-plugin-import';
import jsdoc from 'eslint-plugin-jsdoc';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import tsdoc from 'eslint-plugin-tsdoc';
import prettierConfig from 'eslint-config-prettier';

export default tseslint.config(
  {
    ignores: ['dist/**', 'src/wailsjs/**', 'node_modules/**'],
  },
  {
    files: ['scripts/**/*.mjs'],
    languageOptions: {
      globals: {
        console: 'readonly',
        process: 'readonly',
      },
    },
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['src/**/*.{ts,tsx}'],
    plugins: {
      import: importPlugin,
      jsdoc,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
      tsdoc,
    },
    settings: {
      'import/resolver': {
        node: {
          extensions: ['.js', '.jsx', '.ts', '.tsx'],
        },
      },
      jsdoc: {
        mode: 'typescript',
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
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
    files: ['src/hooks/features/dictionaryBuilder/**/*.{ts,tsx}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      'react-hooks/exhaustive-deps': 'error',
      'max-depth': ['error', 4],
      'no-else-return': ['error', { allowElseIf: false }],
      'no-console': ['warn', { allow: ['warn', 'error'] }],
    },
  },
  {
    files: ['src/hooks/use*.ts', 'src/hooks/features/**/*.{ts,tsx}', 'src/pages/*.tsx'],
    rules: {
      'tsdoc/syntax': 'error',
      'jsdoc/require-jsdoc': [
        'error',
        {
          publicOnly: true,
          contexts: ['TSInterfaceDeclaration', 'TSTypeAliasDeclaration'],
          require: {
            ArrowFunctionExpression: false,
            ClassDeclaration: false,
            ClassExpression: false,
            FunctionDeclaration: true,
            FunctionExpression: false,
            MethodDefinition: false,
          },
        },
      ],
      'jsdoc/require-description': 'error',
      'jsdoc/require-param-description': 'error',
      'jsdoc/require-returns-description': 'error',
      'jsdoc/no-types': 'error',
      'jsdoc/require-param-type': 'off',
      'jsdoc/require-returns-type': 'off',
    },
  },
  {
    files: ['src/pages/*.tsx'],
    rules: {
      'react-refresh/only-export-components': 'off',
    },
  },
  prettierConfig,
);

