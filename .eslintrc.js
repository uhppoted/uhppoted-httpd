module.exports = {
  env: {
    browser: true,
    commonjs: true,
    es2021: true
  },
  extends: [
    'standard'
  ],
  parserOptions: {
    ecmaVersion: 12
  },
  rules: {
    "no-unused-vars": ["error", { "vars": "all", "args": "none", "varsIgnorePattern": "^_" }]
  }
}

