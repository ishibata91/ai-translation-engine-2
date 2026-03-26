import type {MainTranslationCategory} from './types';

const BASE_PROMPTS: Record<MainTranslationCategory, string> = {
    conversation: `You are translating Skyrim dialogue into Japanese.
Preserve speaker intent, tone, and conversational flow.
Use short natural lines that fit in dialogue UI.`,
    quest: `You are translating Skyrim quest text into Japanese.
Preserve quest semantics, objective clarity, and progression context.
Use concise wording that reads well in quest logs.`,
    other: `You are translating Skyrim game text into Japanese.
Preserve meaning, consistency, and in-game readability.`,
};

/**
 * category / recordType から本文翻訳 phase の read-only system prompt を返す。
 */
export const resolveTranslationSystemPrompt = (
    category: MainTranslationCategory,
    recordType: string,
): string => {
    const basePrompt = BASE_PROMPTS[category];
    const normalizedRecordType = recordType.trim() === '' ? 'UNKNOWN' : recordType.trim();

    return `${basePrompt}
Record Type: ${normalizedRecordType}

Requirements:
1. Keep terminology consistent with glossary references when available.
2. Do not add explanations or notes outside the translated text.
3. Prefer natural Japanese phrasing suitable for the selected category.`;
};
