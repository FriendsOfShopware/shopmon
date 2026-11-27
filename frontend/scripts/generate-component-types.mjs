import { glob } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { createServer } from "vite";

const root = fileURLToPath(new URL("..", import.meta.url));
const server = await createServer({
  root,
  configFile: fileURLToPath(new URL("../vite.config.ts", import.meta.url)),
  server: { middlewareMode: true, hmr: false },
  appType: "custom",
  logLevel: "silent",
});

await server.pluginContainer.buildStart({});

const files = [];
for await (const entry of glob("src/**/*.vue", { cwd: root })) {
  files.push(entry);
}

const failures = [];
for (const file of files) {
  try {
    await server.transformRequest(`/${file}`);
  } catch (error) {
    failures.push({ file, error });
  }
}

await server.close();

if (failures.length > 0) {
  for (const { file, error } of failures) {
    console.error(file, error instanceof Error ? error.message : error);
  }
  console.error(`failed to scan ${failures.length} of ${files.length} Vue files`);
  process.exit(1);
}

console.log(`generated component types from ${files.length} Vue files`);
