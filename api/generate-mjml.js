import { readFileSync, readdirSync, writeFileSync } from 'node:fs';
import mjml from 'mjml';

// Directories containing mail templates
const mailDirectories = [
    './src/modules/user/mail',
    './src/modules/organization/mail',
    './src/modules/shop/mail',
];

for (const directory of mailDirectories) {
    const files = readdirSync(directory);

    for (const file of files) {
        if (!file.endsWith('.mjml')) {
            continue;
        }
        const contents = readFileSync(`${directory}/${file}`, 'utf8');
        try {
            const { html } = mjml(contents.toString(), { minify: true });

            writeFileSync(
                `${directory}/${file.replace('.mjml', '.js')}`,
                `export default function(vars) { return \`${html}\`; };`,
            );
        } catch (error) {
            console.error(`Error processing ${file}:`, error);
        }
    }
}
