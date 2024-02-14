import esbuild from "esbuild";
import { existsSync, readFileSync } from "node:fs";
import { execSync } from "node:child_process";
import mjml from "mjml";

// Run the build only once in the CI environment
if (existsSync('worker.api.js') && process.env.CI) {
    console.log('worker.api.js already exists, skipping build');
} else {
    const plugin = {
        name: 'mjml',
        setup(build) {
            build.onLoad({ filter: /\.mjml$/ }, async (args) => {
                const contents = readFileSync(args.path, 'utf8');
                const { html } = mjml(contents, { minify: true });
                return {
                    contents: 'export default function(vars) { return `' + html + '`; };',
                    loader: 'js'
                }
            });
        }
    };

    const commitHash = execSync('git rev-parse HEAD').toString().trim();

    const result = await esbuild.build({
        entryPoints: ['src/index.ts'],
        bundle: true,
        format: 'esm',
        outfile: 'worker.api.js',
        conditions: ['node'],
        plugins: [plugin],
        minify: true,
        sourcemap: true,
        sourceRoot: '/',
        define: {
            'SENTRY_RELEASE': `"${commitHash}"`,
        }
    });

    console.log(result);
}
