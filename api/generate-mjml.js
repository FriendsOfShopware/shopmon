import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import mjml from 'mjml';

const files = readdirSync('./src/mail/sources');

for (const file of files) {
    const contents = readFileSync(`./src/mail/sources/${file}`, 'utf8');
    const { html } = mjml(contents.toString(), { minify: true });

    writeFileSync(
        `./src/mail/sources/${file.replace('.mjml', '.js')}`,
        `export default function(vars) { return \`${html}\`; };`,
    );
}
