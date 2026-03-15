const VAGUE_TOKENS = [
    'might',
    'could',
    'consider',
    'potentially',
    'maybe',
    'perhaps',
    'should',
    'may',
    'possibly',
];
const CONCRETE_PATTERNS = [
    /\b(postgres|redis|mongodb|mysql|elasticsearch|kafka|rabbitmq|nodejs|python|go|rust|java|typescript|javascript|docker|kubernetes|aws|gcp|azure)\b/gi,
    /\d+(\.\d+)?[a-z]*\b/gi,
    /`[^`]+`/g,
    /https?:\/\/\S+/g,
];
export const specificitySc = {
    name: 'Specificity',
    max: 1,
    async score(text, _evalCase) {
        const words = text.toLowerCase().split(/\s+/);
        let vagueCount = 0;
        for (const word of words) {
            for (const vague of VAGUE_TOKENS) {
                if (word.includes(vague)) {
                    vagueCount++;
                    break;
                }
            }
        }
        let concreteCount = 0;
        for (const pattern of CONCRETE_PATTERNS) {
            const matches = text.match(pattern);
            if (matches) {
                concreteCount += matches.length;
            }
        }
        const total = words.length;
        const vagueRatio = vagueCount / total;
        const concreteRatio = concreteCount / Math.max(total, 50);
        // Higher concrete, lower vague = higher score
        const score = Math.max(0, concreteRatio - vagueRatio);
        return Math.min(score, 1);
    },
};
//# sourceMappingURL=specificity.js.map