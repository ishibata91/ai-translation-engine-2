import {render, screen} from '@testing-library/react';
import {describe, expect, it, vi} from 'vitest';
import {DEFAULT_MASTER_PERSONA_LLM_CONFIG, DEFAULT_PERSONA_PROMPT_CONFIG} from '../../types/masterPersona';
import {TerminologyPanel} from './TerminologyPanel';

describe('TerminologyPanel', () => {
    it('translationState が cached の行を翻訳済みテキストとして表示する', () => {
        render(
            <TerminologyPanel
                isActive
                taskId="task-cached"
                summary={{
                    taskId: 'task-cached',
                    status: 'completed',
                    savedCount: 1,
                    failedCount: 1,
                    progressMode: 'hidden',
                    progressCurrent: 0,
                    progressTotal: 0,
                    progressMessage: '',
                }}
                statusLabel="単語翻訳完了"
                errorMessage=""
                targetPage={{
                    taskId: 'task-cached',
                    page: 1,
                    pageSize: 50,
                    totalRows: 2,
                    rows: [
                        {
                            id: 'row-cached',
                            recordType: 'NPC_:FULL',
                            editorId: 'NPC_CACHED',
                            sourceText: 'Cached Source',
                            translatedText: 'キャッシュ翻訳',
                            translationState: 'cached',
                            variant: 'full',
                            sourceFile: 'Update.esm.extract.json',
                        },
                        {
                            id: 'row-missing',
                            recordType: 'NPC_:FULL',
                            editorId: 'NPC_MISSING',
                            sourceText: 'Missing Source',
                            translatedText: '',
                            translationState: 'missing',
                            variant: 'full',
                            sourceFile: 'Update.esm.extract.json',
                        },
                    ],
                }}
                targetStatus="ready"
                targetErrorMessage=""
                isTargetLoading={false}
                isRunning={false}
                llmConfig={{
                    ...DEFAULT_MASTER_PERSONA_LLM_CONFIG,
                    model: 'gemini-2.5-flash',
                }}
                promptConfig={DEFAULT_PERSONA_PROMPT_CONFIG}
                isConfigHydrated
                isPromptHydrated
                onConfigChange={vi.fn()}
                onPromptChange={vi.fn()}
                onRun={vi.fn().mockResolvedValue(undefined)}
                onRefresh={vi.fn().mockResolvedValue(undefined)}
                onTargetPageChange={vi.fn().mockResolvedValue(undefined)}
                onNext={vi.fn()}
            />,
        );

        expect(screen.getByText('キャッシュ翻訳')).toBeTruthy();
        expect(screen.getAllByText('未翻訳')).toHaveLength(1);
    });
});
