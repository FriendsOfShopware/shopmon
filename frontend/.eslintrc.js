module.exports = {
    env: {
        browser: true,
        es2021: true,
        node: true,
    },
    extends: [
        'eslint:recommended',
        'plugin:vue/vue3-recommended',
        'plugin:@typescript-eslint/recommended',
    ],
    parser: 'vue-eslint-parser',
    parserOptions: {
        parser: '@typescript-eslint/parser',
        sourceType: 'module',
    },
    settings: {
        'import/no-useless-path-segments': 0,
        'import/resolver': {
            typescript: {},
        },
    },
    rules: { 
        'vue/component-name-in-template-casing': ['error', 'kebab-case'],
        quotes: ['error', 'single', { 'avoidEscape': true }],
        'vue/no-boolean-default': ['error', 'default-false'],
        'vue/prefer-true-attribute-shorthand': ['error'],
        'vue/no-multiple-objects-in-class': ['error'],
        'comma-dangle': ['error', 'always-multiline'],
        'vue/padding-line-between-blocks': ['error'],
        'vue/next-tick-style': ['error', 'promise' ],
        'vue/v-for-delimiter-style': ['error', 'in'],
        'vue/html-button-has-type': ['error'],
        'vue/multi-word-component-names': 0,
        'vue/html-indent': ['error', 4],
        'semi': ['error', 'always'],
        'max-len': ['error', 125],
        'no-extra-semi': 'error',
        indent: ['error', 4],
        'no-undef': 0,
    },
};
