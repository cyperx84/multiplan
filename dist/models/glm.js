import { readFileSync } from 'fs';
import { homedir } from 'os';
import { join } from 'path';
import https from 'https';
import { URL } from 'url';
export class GLMAdapter {
    id = 'glm5';
    name = 'GLM-5 (ZhipuAI)';
    apiUrl = 'https://api.z.ai/api/coding/paas/v4/chat/completions';
    async available() {
        try {
            this.getApiKey();
            return true;
        }
        catch {
            return false;
        }
    }
    async plan(prompt, timeoutMs = 120000) {
        const key = this.getApiKey();
        const payload = JSON.stringify({
            model: 'glm-5',
            messages: [{ role: 'user', content: prompt }],
            max_tokens: 8192,
            temperature: 0.7,
        });
        return new Promise((resolve, reject) => {
            const url = new URL(this.apiUrl);
            const request = https.request({
                hostname: url.hostname,
                path: url.pathname + url.search,
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${key}`,
                    'Content-Type': 'application/json',
                    'Content-Length': Buffer.byteLength(payload),
                },
                timeout: timeoutMs,
            }, (res) => {
                let data = '';
                res.on('data', (chunk) => {
                    data += chunk;
                });
                res.on('end', () => {
                    try {
                        const json = JSON.parse(data);
                        if (!json.choices || !json.choices[0]?.message?.content) {
                            reject(new Error(`Unexpected GLM-5 response: ${data.substring(0, 200)}`));
                        }
                        else {
                            resolve(json.choices[0].message.content);
                        }
                    }
                    catch (err) {
                        reject(new Error(`GLM-5 parse error: ${data.substring(0, 200)}`));
                    }
                });
            });
            request.on('error', (err) => {
                reject(new Error(`GLM-5 request failed: ${err.message}`));
            });
            request.on('timeout', () => {
                request.destroy();
                reject(new Error(`GLM-5 timed out after ${timeoutMs}ms`));
            });
            request.write(payload);
            request.end();
        });
    }
    getApiKey() {
        // 1. Environment variable
        const envKey = process.env.ZAI_API_KEY || process.env.GLM_API_KEY;
        if (envKey) {
            return envKey;
        }
        // 2. OpenClaw auth-profiles.json
        const paths = [
            join(homedir(), '.openclaw/agents/main/agent/auth-profiles.json'),
            join(homedir(), '.openclaw/agents/builder/agent/auth-profiles.json'),
        ];
        for (const path of paths) {
            try {
                const content = readFileSync(path, 'utf-8');
                const data = JSON.parse(content);
                const profile = data.profiles?.['zai:default'];
                if (profile?.key) {
                    return profile.key;
                }
            }
            catch {
                // continue to next path
            }
        }
        throw new Error('ZAI API key not found. Set ZAI_API_KEY env var or ensure ~/.openclaw/agents/main/agent/auth-profiles.json has zai:default profile.');
    }
}
//# sourceMappingURL=glm.js.map