module.exports = {
    root: true,
    env: {
        browser: true,
        es2021: true,
        node: true
    },
    extends: [
        'plugin:@typescript-eslint/recommended',
    ],
    parserOptions: {
        parser: "@typescript-eslint/parser",
        sourceType: "module",
    },
    settings: {
        'import/resolver': {
            typescript: {}
        }
    },
    rules: { 
        'no-multiple-empty-lines': ['error', { max: 2, maxEOF: 1 }],
        'object-curly-spacing': ["error", "always"],
        indent: ['error', 4],
        'no-undef': ['off'],
    },
}
