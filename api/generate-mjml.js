import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import mjml from 'mjml';

const files = readdirSync('./src/mail/sources');

for (const file of files) {
    if (!file.endsWith('.mjml')) {
        continue;
    }
    const contents = readFileSync(`./src/mail/sources/${file}`, 'utf8');
    try {
        const { html } = mjml(contents.toString(), { minify: true });

        writeFileSync(
            `./src/mail/sources/${file.replace('.mjml', '.js')}`,
            `export default function(vars) { return \`${html}\`; };`,
        );
    } catch (error) {
        console.error(`Error processing ${file}:`, error);
    }
}
