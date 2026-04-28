const js = require("@eslint/js");

module.exports = [
  // 1. Apply recommended rules to all files
  js.configs.recommended,

  // 2. BROWSER CONFIG: Only for files in the "web" folder
  {
    files: ["web/**/*.js"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "script",
      globals: {
        HTMLElement: "readonly", // 👈 Whitelisted just this one as requested
        console: "readonly",
        document: "readonly",
        fetch: "readonly",
        URL: "readonly",
        window: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        requestAnimationFrame: "readonly",
      },
    },
    rules: {
      "no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
    },
  },

  // 3. NODE.JS CONFIG: For files outside the "web" folder (like build scripts)
  {
    files: ["**/*.js"],
    ignores: ["web/**/*.js"], // 👈 Prevents Node globals from leaking into web files
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "commonjs", // 👈 Tells ESLint you are using 'require' and 'module.exports'
      globals: {
        __dirname: "readonly",
        __filename: "readonly",
        console: "readonly",
        module: "readonly",
        process: "readonly",
        require: "readonly",
      },
    },
  },
];
