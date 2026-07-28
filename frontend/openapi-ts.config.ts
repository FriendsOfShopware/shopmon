import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../api/openapi/spec.yaml",
  output: "src/api/generated",
  plugins: ["@hey-api/client-fetch"],
});
