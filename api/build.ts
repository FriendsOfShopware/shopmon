import { rm, cp } from "node:fs/promises";
import { join } from "node:path";

const outdir = "dist";

// Clean the output directory
await rm(outdir, { recursive: true, force: true });

// Build all entry points
const entrypoints = ["app.ts", "cron.ts", "migrate.ts"];

console.log("Building entrypoints:", entrypoints.join(", "));

const result = await Bun.build({
  entrypoints,
  outdir,
  target: "bun",
  sourcemap: "external",
  bytecode: true,
  minify: {
    whitespace: true,
    syntax: true,
    identifiers: false,
  },
});

if (!result.success) {
  console.error("Build failed:");
  for (const log of result.logs) {
    console.error(log);
  }
  process.exit(1);
}

console.log(`Built ${result.outputs.length} files`);

// Copy drizzle migrations
const drizzleSrc = "drizzle";
const drizzleDest = join(outdir, "drizzle");

await cp(drizzleSrc, drizzleDest, { recursive: true });
console.log("Copied drizzle migrations to dist/drizzle");

console.log("Build complete!");
