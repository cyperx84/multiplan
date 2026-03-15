import { spawn } from 'child_process';
export class GeminiAdapter {
    id = 'gemini';
    name = 'Gemini';
    async available() {
        return this.commandExists('gemini');
    }
    async plan(prompt, timeoutMs = 120000) {
        return this.runCommand('gemini', ['-p', prompt], timeoutMs);
    }
    commandExists(cmd) {
        return new Promise((resolve) => {
            const proc = spawn('sh', ['-c', `command -v ${cmd}`]);
            let code = null;
            proc.on('close', (exitCode) => {
                code = exitCode;
                resolve(code === 0);
            });
            setTimeout(() => {
                if (code === null) {
                    proc.kill();
                    resolve(false);
                }
            }, 1000);
        });
    }
    runCommand(cmd, args, timeoutMs) {
        return new Promise((resolve, reject) => {
            const proc = spawn(cmd, args);
            let stdout = '';
            let stderr = '';
            let timedOut = false;
            const timeout = setTimeout(() => {
                timedOut = true;
                proc.kill();
                reject(new Error(`${cmd} timed out after ${timeoutMs}ms`));
            }, timeoutMs);
            proc.stdout?.on('data', (data) => {
                stdout += data.toString();
            });
            proc.stderr?.on('data', (data) => {
                stderr += data.toString();
            });
            proc.on('close', (code) => {
                clearTimeout(timeout);
                if (timedOut)
                    return;
                if (code !== 0) {
                    reject(new Error(`${cmd} exited with code ${code}: ${stderr}`));
                }
                else {
                    resolve(stdout);
                }
            });
            proc.on('error', (err) => {
                clearTimeout(timeout);
                reject(err);
            });
        });
    }
}
//# sourceMappingURL=gemini.js.map