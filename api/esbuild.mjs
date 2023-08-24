import esbuild from "esbuild";
import { existsSync } from "node:fs";
import { execSync } from "node:child_process";

// Run the build only once in the CI environment
if (existsSync('worker.api.js') && process.env.CI) {
    console.log('worker.api.js already exists, skipping build');
} else {
    const commitHash = execSync('git rev-parse HEAD').toString().trim();

    const result = await esbuild.build({
        entryPoints: ['src/index.ts'],
        bundle: true,
        format: 'esm',
        outfile: 'worker.api.js',
        conditions: ['node'],
        minify: true,
        sourcemap: true,
        sourceRoot: '/',
        define: {
            'SENTRY_RELEASE': `"${commitHash}"`,
        }
    });
    
    console.log(result);
}
