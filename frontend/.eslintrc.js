module.exports = {
    env: {
        browser: true,
        es2021: true,
        node: true
    },
    extends: [
        'eslint:recommended',
        'plugin:vue/vue3-recommended',
        'plugin:@typescript-eslint/recommended',
    ],
    parser: "vue-eslint-parser",
    parserOptions: {
        parser: "@typescript-eslint/parser",
        sourceType: "module"
    },
    settings: {
        'import/resolver': {
            typescript: {}
        }
    },
    rules: { 
        indent: ['error', 4],
        'vue/multi-word-component-names': 0,
        'no-undef': 0
    },
}