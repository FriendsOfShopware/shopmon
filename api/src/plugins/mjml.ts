import { plugin } from 'bun';
import mjml from 'mjml';

plugin({
    name: 'mjml-loader',
    setup(build) {
        build.onLoad({ filter: /\.mjml$/ }, async (args) => {
            const text = await Bun.file(args.path).text();
            const { html } = mjml(text, { minify: true });

            return {
                contents: `export default function(vars) { return 
${html}
};`,
                loader: 'js',
            };
        });
    },
});
