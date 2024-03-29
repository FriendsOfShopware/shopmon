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
    plugins: [
        '@stylistic',
    ],
    rules: {
        '@stylistic/quotes': ['error', 'single', { 'avoidEscape': true }],
        '@stylistic/comma-dangle': ['error', 'always-multiline'],
        '@stylistic/semi': ['error', 'always'],
        '@stylistic/max-len': ['error', 125],
        '@stylistic/no-extra-semi': 'error',
        '@stylistic/indent': ['error', 4],
        '@stylistic/no-trailing-spaces': ['error'],
        '@stylistic/key-spacing': ['error'],
        '@stylistic/object-curly-spacing': ['error', 'always'],
        '@stylistic/space-before-function-paren': ['error', 'never'],
        '@stylistic/space-infix-ops': ['error'],
        '@stylistic/space-unary-ops': ['error'],
        '@stylistic/arrow-spacing': ['error'],
        '@stylistic/no-multi-spaces': ['error'],
        '@stylistic/keyword-spacing': ['error'],
        'vue/component-name-in-template-casing': ['error', 'kebab-case'],
        'vue/no-boolean-default': ['error', 'default-false'],
        'vue/prefer-true-attribute-shorthand': ['error'],
        'vue/no-multiple-objects-in-class': ['error'],
        'vue/padding-line-between-blocks': ['error'],
        'vue/next-tick-style': ['error', 'promise' ],
        'vue/v-for-delimiter-style': ['error', 'in'],
        'vue/html-button-has-type': ['error'],
        'vue/multi-word-component-names': 0,
        'vue/html-indent': ['error', 4],
        'vue/block-order': ['error', {
            'order': [  'template', 'script', 'style' ],
        }],
        'no-undef': 0,
    },
};
