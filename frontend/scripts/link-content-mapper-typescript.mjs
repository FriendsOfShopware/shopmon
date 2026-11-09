import { mkdirSync, rmSync, symlinkSync } from "node:fs";
import { dirname, resolve } from "node:path";

// @vue/content-mapper imports typescript/sync, which exists only on TypeScript 7.
// Vue's compiler still needs the classic API, so the `typescript` package stays on 6.x
// and this link exposes the 7.1 alias to the mapper. The package.json peer
// override lets npm resolve the root 6.x package; this link supplies the actual
// 7.x runtime required by the mapper after installation.
const link = resolve("node_modules/@vue/content-mapper/node_modules/typescript");
mkdirSync(dirname(link), { recursive: true });
rmSync(link, { recursive: true, force: true });
symlinkSync("../../../typescript-7", link, "dir");
