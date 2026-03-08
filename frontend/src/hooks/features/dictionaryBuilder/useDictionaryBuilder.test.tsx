import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { useDictionaryBuilder } from './useDictionaryBuilder';
import * as AppBindings from '../../../wailsjs/go/main/App';

const eventHandlers = new Map<string, (payload: unknown) => void>();

type MockFnLike = {
    mockResolvedValue: (value: unknown) => void;
};

const asMock = (fn: unknown): MockFnLike => fn as MockFnLike;

vi.mock('../../../wailsjs/runtime/runtime', () => ({
    EventsOn: vi.fn((eventName: string, callback: (payload: unknown) => void) => {
        eventHandlers.set(eventName, callback);
        return () => {
            eventHandlers.delete(eventName);
        };
    }),
}));

vi.mock('../../../wailsjs/go/main/App', () => ({
    DictGetSources: vi.fn(),
    DictStartImport: vi.fn(),
    DictGetEntriesPaginated: vi.fn(),
    DictSearchAllEntriesPaginated: vi.fn(),
    DictUpdateEntry: vi.fn(),
    DictDeleteEntry: vi.fn(),
    SelectFiles: vi.fn(),
    DictDeleteSource: vi.fn(),
}));

describe('useDictionaryBuilder', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        eventHandlers.clear();

        asMock(AppBindings.DictGetSources).mockResolvedValue([
            {
                id: 1,
                file_name: 'Skyrim.esm.xml',
                format: 'sstxml',
                entry_count: 12,
                status: 'COMPLETED',
                imported_at: '2026-03-08T01:00:00Z',
                file_path: 'C:/tmp/Skyrim.esm.xml',
                file_size_bytes: 2048,
                error_message: undefined,
            },
        ]);
        asMock(AppBindings.DictGetEntriesPaginated).mockResolvedValue({ entries: [], totalCount: 0 });
        asMock(AppBindings.DictSearchAllEntriesPaginated).mockResolvedValue({ entries: [], totalCount: 0 });
        asMock(AppBindings.SelectFiles).mockResolvedValue([]);
        asMock(AppBindings.DictDeleteEntry).mockResolvedValue(undefined);
        asMock(AppBindings.DictUpdateEntry).mockResolvedValue(undefined);
        asMock(AppBindings.DictDeleteSource).mockResolvedValue(undefined);
        asMock(AppBindings.DictStartImport).mockResolvedValue(1);
    });

    it('初期化時にソース一覧を取得して state に反映する', async () => {
        const { result } = renderHook(() => useDictionaryBuilder());

        await waitFor(() => {
            expect(result.current.state.sources).toHaveLength(1);
        });

        const source = result.current.state.sources[0];
        expect(source.fileName).toBe('Skyrim.esm.xml');
        expect(source.status).toBe('完了');
        expect(result.current.state.view).toBe('list');
    });

    it('import_progress イベントで進捗状態を更新する', async () => {
        const { result } = renderHook(() => useDictionaryBuilder());

        await waitFor(() => {
            expect(eventHandlers.has('dictionary:import_progress')).toBe(true);
        });

        const handler = eventHandlers.get('dictionary:import_progress');
        expect(handler).toBeDefined();

        act(() => {
            handler?.({
                CorrelationID: 'corr-1',
                Status: 'IN_PROGRESS',
                Message: '解析中',
                Total: 10,
                Completed: 3,
            });
        });

        expect(result.current.state.isImporting).toBe(true);
        expect(result.current.state.importMessages['corr-1']).toBe('解析中');

        act(() => {
            handler?.({
                CorrelationID: 'corr-1',
                Status: 'COMPLETED',
                Message: '完了',
                Total: 10,
                Completed: 10,
            });
        });

        await waitFor(() => {
            expect(result.current.state.importMessages['corr-1']).toBeUndefined();
        });
        expect(result.current.state.isImporting).toBe(false);
    });
});
