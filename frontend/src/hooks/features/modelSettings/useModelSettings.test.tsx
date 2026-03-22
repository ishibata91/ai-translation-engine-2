import {renderHook, waitFor} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {useModelSettings} from './useModelSettings';
import {type MasterPersonaLLMConfig} from '../../../types/masterPersona';
import * as ModelCatalogBindings from '../../../wailsjs/go/controller/ModelCatalogController';
import {modelcatalog} from '../../../wailsjs/go/models';

vi.mock('../../../wailsjs/go/controller/ModelCatalogController', () => ({
    ListModels: vi.fn(),
}));

const createConfig = (overrides: Partial<MasterPersonaLLMConfig> = {}): MasterPersonaLLMConfig => ({
    provider: 'gemini',
    model: 'gemini-3.1-flash-lite-preview',
    endpoint: '',
    apiKey: 'test-key',
    temperature: 0.3,
    contextLength: 0,
    syncConcurrency: 1,
    bulkStrategy: 'sync',
    ...overrides,
});

const createCatalogModel = (overrides: Partial<modelcatalog.ModelOption> = {}): modelcatalog.ModelOption => modelcatalog.ModelOption.createFrom({
    id: 'models/gemini-3.1-flash-lite-preview',
    display_name: 'Gemini 3.1 Flash Lite Preview',
    loaded: false,
    capability: modelcatalog.ModelCapability.createFrom({ supports_batch: true }),
    ...overrides,
});

describe('useModelSettings', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('catalog 読み込み前でも保存済み Gemini モデルを選択状態のまま維持する', async () => {
        const onChange = vi.fn();
        vi.mocked(ModelCatalogBindings.ListModels).mockReturnValue(new Promise(() => undefined));

        const {result} = renderHook(() => useModelSettings({
            value: createConfig(),
            onChange,
            enabled: true,
            namespace: 'translation_flow.terminology',
        }));

        await waitFor(() => {
            expect(ModelCatalogBindings.ListModels).toHaveBeenCalled();
        });

        expect(result.current.selectedModelValue).toBe('gemini-3.1-flash-lite-preview');
        expect(result.current.selectableModelOptions[0]?.id).toBe('gemini-3.1-flash-lite-preview');
        expect(onChange).not.toHaveBeenCalled();
    });

    it('Gemini catalog の models/ 接頭辞を正規化して保存済み値と一致扱いにする', async () => {
        const onChange = vi.fn();
        vi.mocked(ModelCatalogBindings.ListModels).mockResolvedValue([
            createCatalogModel(),
        ]);

        const {result} = renderHook(() => useModelSettings({
            value: createConfig(),
            onChange,
            enabled: true,
            namespace: 'translation_flow.terminology',
        }));

        await waitFor(() => {
            expect(result.current.selectableModelOptions).toEqual([
                {
                    id: 'gemini-3.1-flash-lite-preview',
                    label: 'Gemini 3.1 Flash Lite Preview',
                    capability: {supportsBatch: true},
                },
            ]);
        });

        expect(result.current.selectedModelValue).toBe('gemini-3.1-flash-lite-preview');
        expect(onChange).not.toHaveBeenCalled();
    });

    it('catalog に保存済みモデルが存在しない場合だけ先頭候補へ切り替える', async () => {
        const onChange = vi.fn();
        vi.mocked(ModelCatalogBindings.ListModels).mockResolvedValue([
            createCatalogModel(),
        ]);

        renderHook(() => useModelSettings({
            value: createConfig({model: 'missing-model'}),
            onChange,
            enabled: true,
            namespace: 'translation_flow.terminology',
        }));

        await waitFor(() => {
            expect(onChange).toHaveBeenCalledWith(expect.objectContaining({
                model: 'gemini-3.1-flash-lite-preview',
            }));
        });
    });
});
