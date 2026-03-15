const REQUIRED_SECTIONS = [
    'Overview',
    'Architecture',
    'Implementation',
    'Trade-off',
    'Risk',
];
export const coverageScorer = {
    name: 'Coverage',
    max: 1,
    async score(text, _evalCase) {
        let found = 0;
        for (const section of REQUIRED_SECTIONS) {
            if (new RegExp(`##\\s+${section}`, 'i').test(text)) {
                found++;
            }
        }
        return Math.min(found / REQUIRED_SECTIONS.length, 1);
    },
};
//# sourceMappingURL=coverage.js.map