import js from "@eslint/js";
import typescript from "@typescript-eslint/eslint-plugin";
import typescriptParser from "@typescript-eslint/parser";
import vue from "eslint-plugin-vue";
import vueParser from "vue-eslint-parser";

export default [
  // JavaScript base configuration
  js.configs.recommended,

  // Vue configurations
  ...vue.configs["flat/essential"],
  ...vue.configs["flat/strongly-recommended"],
  ...vue.configs["flat/recommended"],

  // Global ignores
  {
    ignores: [
      "dist/**",
      "node_modules/**",
      "*.config.js",
      "*.config.mjs",
      "*.config.ts",
      ".vite/**",
      "components.d.ts", // Auto-generated types
      "auto-imports.d.ts", // Auto-generated imports
    ],
  },

  // TypeScript and Vue files configuration
  {
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: typescriptParser,
        ecmaVersion: "latest",
        sourceType: "module",
        extraFileExtensions: [".vue"],
        project: "./tsconfig.json",
        tsconfigRootDir: import.meta.dirname,
      },
      globals: {
        // Browser globals
        window: "readonly",
        document: "readonly",
        navigator: "readonly",
        console: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        localStorage: "readonly",
        sessionStorage: "readonly",
        crypto: "readonly",
        TextEncoder: "readonly",
        TextDecoder: "readonly",
        URL: "readonly",
        URLSearchParams: "readonly",
        // Node globals for config files
        process: "readonly",
        __dirname: "readonly",
        __filename: "readonly",
        Buffer: "readonly",
        global: "readonly",
        // Vue globals
        defineProps: "readonly",
        defineEmits: "readonly",
        defineExpose: "readonly",
        withDefaults: "readonly",
      },
    },
    plugins: {
      "@typescript-eslint": typescript,
      vue,
    },
    rules: {
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/explicit-function-return-type": "off",
      "@typescript-eslint/explicit-module-boundary-types": "off",
      "@typescript-eslint/no-non-null-assertion": "warn",

      // Use TypeScript-aware unused vars rule
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "no-unused-vars": "off",

      // Type-aware rules (require TypeScript compilation)
      "@typescript-eslint/no-floating-promises": "off",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/await-thenable": "error",
      "@typescript-eslint/no-unnecessary-type-assertion": "error",
      "@typescript-eslint/prefer-nullish-coalescing": "warn",
      "@typescript-eslint/prefer-optional-chain": "warn",
      "@typescript-eslint/strict-boolean-expressions": "off", // Too strict for Vue templates

      // Vue rules
      "vue/multi-word-component-names": "off",
      "vue/no-v-html": "warn",
      "vue/component-definition-name-casing": ["error", "PascalCase"],
      "vue/component-name-in-template-casing": "off", // Allow both kebab-case and PascalCase
      "vue/prop-name-casing": ["error", "camelCase"],
      "vue/attribute-hyphenation": "off", // Allow both styles
      "vue/v-on-event-hyphenation": "off", // Allow both styles
      "vue/max-attributes-per-line": "off", // Too strict for existing code
      "vue/first-attribute-linebreak": "off", // Too strict for existing code
      "vue/html-closing-bracket-newline": "off", // Too strict for existing code
      "vue/html-indent": "off", // Project uses 4-space indentation consistently
      "vue/script-indent": "off", // Project uses 4-space indentation consistently
      "vue/no-mutating-props": "warn", // Warning instead of error
      "vue/singleline-html-element-content-newline": "off", // Too strict
      "vue/attributes-order": "warn", // Warning instead of error
      "vue/html-self-closing": "off", // Conflicts with oxfmt formatter

      // General rules
      "no-console": "warn",
      "no-debugger": "error",
      "prefer-const": "error",
      "no-var": "error",

      // Disable no-useless-assignment for Vue files
      // This rule doesn't understand Vue's <script setup> pattern where
      // variables are exposed to templates automatically
      "no-useless-assignment": "off",
    },
  },

  // JavaScript files configuration
  {
    files: ["**/*.js"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
    },
    rules: {
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
      "no-console": "warn",
      "no-debugger": "error",
      "prefer-const": "error",
      "no-var": "error",
    },
  },

  // Test files configuration
  {
    files: ["**/*.spec.ts"],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: typescriptParser,
        ecmaVersion: "latest",
        sourceType: "module",
        extraFileExtensions: [".vue"],
        project: "./tsconfig.json",
        tsconfigRootDir: import.meta.dirname,
      },
      globals: {
        // Vitest globals
        describe: "readonly",
        it: "readonly",
        expect: "readonly",
        vi: "readonly",
        beforeEach: "readonly",
        afterEach: "readonly",
        beforeAll: "readonly",
        afterAll: "readonly",
        // Browser globals
        window: "readonly",
        document: "readonly",
        console: "readonly",
        // Vue globals
        defineProps: "readonly",
        defineEmits: "readonly",
        defineExpose: "readonly",
      },
    },
    plugins: {
      "@typescript-eslint": typescript,
      vue,
    },
    rules: {
      // Allow multiple components in test files (for stubs)
      "vue/one-component-per-file": "off",
      // Allow test stubs without prop types
      "vue/require-prop-types": "off",
      // Allow any in test mocks
      "@typescript-eslint/no-explicit-any": "off",
      // General rules
      "no-console": "off",
    },
  },
];
