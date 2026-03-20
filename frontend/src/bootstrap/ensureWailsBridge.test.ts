import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';

type RuntimeMethod = (...args: unknown[]) => unknown;
type ControllerMethod = (...args: unknown[]) => Promise<unknown>;

type TestWindow = Window & {
    runtime?: Record<string, RuntimeMethod>;
    go?: {
        controller?: Record<string, Record<string, ControllerMethod>>;
    };
};

describe('ensureWailsBridge', () => {
    let ensureWailsBridge: () => void;
    let win: TestWindow;

    beforeEach(async () => {
        vi.resetModules();
        ({ensureWailsBridge} = await import('./ensureWailsBridge'));
        win = window as TestWindow;
        delete win.runtime;
        delete win.go;
        window.location.hash = '';
    });

    afterEach(() => {
        vi.restoreAllMocks();
        delete win.runtime;
        delete win.go;
        window.location.hash = '';
    });

    it('Wails未注入時に no-op bridge を補完する', async () => {
        const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => undefined);
        window.location.hash = '#/translation_flow';

        ensureWailsBridge();

        expect(typeof win.runtime?.EventsOnMultiple).toBe('function');
        const unsubscribe = win.runtime?.EventsOnMultiple('event', () => undefined, -1);
        expect(typeof unsubscribe).toBe('function');

        const configGetAll = win.go?.controller?.ConfigController?.ConfigGetAll;
        await expect(configGetAll?.('translation_flow.terminology.llm')).resolves.toEqual({});

        const taskGetAll = win.go?.controller?.TaskController?.GetAllTasks;
        const taskList = (await taskGetAll?.()) as Array<{id: string; type: string}>;
        expect(taskList.length).toBeGreaterThan(0);
        expect(taskList[0]?.type).toBe('translation_project');

        const listLoadedFiles = win.go?.controller?.TaskController?.ListLoadedTranslationFlowFiles;
        const loadResult = (await listLoadedFiles?.(taskList[0]?.id ?? '')) as {
            task_id: string;
            files: Array<{file_id: number}>;
        };
        expect(loadResult.files.length).toBeGreaterThan(0);
        expect(loadResult.files[0]?.file_id).toBeGreaterThan(0);

        const unknownMethod = win.go?.controller?.UnknownController?.UnknownMethod;
        await expect(unknownMethod?.()).resolves.toBeUndefined();

        expect(infoSpy).toHaveBeenCalledTimes(1);
    });

    it('既存の runtime/controller 注入は上書きしない', async () => {
        const customUnsubscribe = () => undefined;
        const customEventsOnMultiple = vi.fn(() => customUnsubscribe);
        const customConfigGetAll = vi.fn(async () => ({provider: 'gemini'}));
        win.runtime = {
            EventsOnMultiple: customEventsOnMultiple,
        };
        win.go = {
            controller: {
                ConfigController: {
                    ConfigGetAll: customConfigGetAll,
                },
            },
        };

        ensureWailsBridge();

        expect(win.runtime.EventsOnMultiple).toBe(customEventsOnMultiple);
        expect(win.runtime.EventsOnMultiple('event', () => undefined, -1)).toBe(customUnsubscribe);
        const configGetAll = win.go?.controller?.ConfigController?.ConfigGetAll;
        await expect(configGetAll?.('translation_flow.terminology.llm')).resolves.toEqual({
            provider: 'gemini',
        });
    });

    it('再実行しても shim の通知ログは1回のみ', () => {
        const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => undefined);

        ensureWailsBridge();
        ensureWailsBridge();

        expect(infoSpy).toHaveBeenCalledTimes(1);
    });
});
