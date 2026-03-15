export interface PromptVariables {
    [key: string]: string;
}
export declare function loadPrompt(name: string): string;
export declare function renderPrompt(template: string, variables: PromptVariables): string;
export declare function loadAndRender(name: string, variables: PromptVariables): string;
//# sourceMappingURL=loader.d.ts.map