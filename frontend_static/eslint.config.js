// eslint.config.js
import js from '@eslint/js';
import globals from 'globals';

export default [
  js.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    },
    rules: {
      "semi": "error",
      "no-unused-vars": "error",
      "no-var": "error",
      "no-undef": "error"
    }
  }
];