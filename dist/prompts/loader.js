import { readFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
const __dirname = dirname(fileURLToPath(import.meta.url));
export function loadPrompt(name) {
    const path = join(__dirname, `${name}.md`);
    return readFileSync(path, 'utf-8');
}
export function renderPrompt(template, variables) {
    let result = template;
    for (const [key, value] of Object.entries(variables)) {
        const placeholder = `{{${key}}}`;
        result = result.replace(new RegExp(placeholder, 'g'), value);
    }
    return result;
}
export function loadAndRender(name, variables) {
    const template = loadPrompt(name);
    return renderPrompt(template, variables);
}
//# sourceMappingURL=loader.js.map