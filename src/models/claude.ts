import { spawn } from 'child_process';
import { ModelAdapter } from './types.js';

export class ClaudeAdapter implements ModelAdapter {
  id = 'claude';
  name = 'Claude (Opus)';

  async available(): Promise<boolean> {
    return this.commandExists('claude');
  }

  async plan(prompt: string, timeoutMs = 120000): Promise<string> {
    return this.runCommand(
      'claude',
      ['--print', '--permission-mode', 'bypassPermissions'],
      prompt,
      timeoutMs
    );
  }

  private commandExists(cmd: string): Promise<boolean> {
    return new Promise((resolve) => {
      const proc = spawn('sh', ['-c', `command -v ${cmd}`]);
      let code: number | null = null;

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

  private runCommand(
    cmd: string,
    args: string[],
    stdin: string,
    timeoutMs: number
  ): Promise<string> {
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
        if (timedOut) return;

        if (code !== 0) {
          reject(new Error(`${cmd} exited with code ${code}: ${stderr}`));
        } else {
          resolve(stdout);
        }
      });

      proc.on('error', (err) => {
        clearTimeout(timeout);
        reject(err);
      });

      proc.stdin?.write(stdin);
      proc.stdin?.end();
    });
  }
}
