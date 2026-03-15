import {useCallback, useEffect, useMemo, useState} from 'react';
import {useLocation} from 'react-router-dom';
import {SelectTranslationInputFiles} from '../../../wailsjs/go/controller/FileDialogController';
import {
    GetAllTasks,
    ListLoadedTranslationFlowFiles,
    ListTranslationFlowPreviewRows,
    LoadTranslationFlowFiles,
} from '../../../wailsjs/go/controller/TaskController';
import {mapLoadResult, mapPreviewPage} from './adapters';
import type {LoadedTranslationFile, TranslationFlowTab, UseTranslationFlowResult} from './types';

const PREVIEW_PAGE_SIZE = 50;

const TABS: TranslationFlowTab[] = [
    {label: 'データロード'},
    {label: '用語'},
    {label: 'ペルソナ生成'},
    {label: '要約'},
    {label: '翻訳'},
    {label: 'エクスポート'},
];

const TO_TASK_RESOLVE_ERROR = 'translation_project task の取得に失敗しました';

const toErrorMessage = (error: unknown, fallback: string): string => {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message;
    }
    if (typeof error === 'string' && error.trim() !== '') {
        return error;
    }
    if (error && typeof error === 'object') {
        const maybeMessage = (error as {message?: unknown}).message;
        if (typeof maybeMessage === 'string' && maybeMessage.trim() !== '') {
            return maybeMessage;
        }
    }
    return fallback;
};

const parseDate = (value: unknown): number => {
    if (typeof value !== 'string' || value.trim() === '') {
        return 0;
    }
    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? 0 : parsed;
};

const resolveTranslationProjectTaskID = (payload: unknown): string => {
    if (!Array.isArray(payload)) {
        return '';
    }

    const candidates = payload
        .filter((entry): entry is Record<string, unknown> => Boolean(entry) && typeof entry === 'object')
        .map((entry) => ({
            id: typeof entry.id === 'string' ? entry.id.trim() : '',
            type: typeof entry.type === 'string' ? entry.type : '',
            status: typeof entry.status === 'string' ? entry.status : '',
            updatedAt: parseDate(entry.updated_at ?? entry.updatedAt),
        }))
        .filter((entry) => entry.type === 'translation_project' && entry.id !== '');

    if (candidates.length === 0) {
        return '';
    }

    const active = candidates
        .filter((entry) => entry.status !== 'completed')
        .sort((a, b) => b.updatedAt - a.updatedAt);

    if (active.length > 0) {
        return active[0].id;
    }

    const sorted = [...candidates].sort((a, b) => b.updatedAt - a.updatedAt);
    return sorted[0]?.id ?? '';
};

const toTaskIDFromRoute = (value: unknown): string => {
    if (typeof value !== 'string') {
        return '';
    }
    return value.trim();
};

const resolveTaskID = async (routeTaskID: string): Promise<string> => {
    if (routeTaskID !== '') {
        return routeTaskID;
    }

    const allTasks = await GetAllTasks();
    return resolveTranslationProjectTaskID(allTasks);
};

/**
 * TranslationFlow 画面のロードフェーズ状態を headless に管理する。
 */
export function useTranslationFlow(): UseTranslationFlowResult {
    const location = useLocation();
    const navState = location.state as {taskId?: string} | null;

    const routeTaskID = useMemo(() => toTaskIDFromRoute(navState?.taskId), [navState?.taskId]);

    const [taskId, setTaskID] = useState(routeTaskID);
    const [activeTab, setActiveTab] = useState(0);
    const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
    const [loadedFiles, setLoadedFiles] = useState<LoadedTranslationFile[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');

    useEffect(() => {
        let active = true;

        const run = async () => {
            try {
                const resolved = await resolveTaskID(routeTaskID);
                if (!active) {
                    return;
                }
                setTaskID(resolved);
                if (resolved === '') {
                    setErrorMessage('');
                    return;
                }
                setErrorMessage('');
            } catch (error) {
                if (!active) {
                    return;
                }
                setErrorMessage(toErrorMessage(error, TO_TASK_RESOLVE_ERROR));
            }
        };

        void run();

        return () => {
            active = false;
        };
    }, [routeTaskID]);

    const mergeUniquePaths = (base: string[], incoming: string[]): string[] => {
        const existing = new Set(base);
        const next = [...base];
        for (const path of incoming) {
            if (!existing.has(path)) {
                existing.add(path);
                next.push(path);
            }
        }
        return next;
    };

    const handleReloadFiles = useCallback(async () => {
        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await ListLoadedTranslationFlowFiles(taskId);
            const mapped = mapLoadResult(payload);
            if (mapped.taskId !== '' && mapped.taskId !== taskId) {
                setTaskID(mapped.taskId);
            }
            setLoadedFiles(mapped.files);
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ロード済みファイルの取得に失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, [taskId]);

    useEffect(() => {
        void handleReloadFiles();
    }, [handleReloadFiles]);

    const handleSelectFiles = useCallback(async () => {
        setErrorMessage('');
        try {
            const files = await SelectTranslationInputFiles();
            if (!Array.isArray(files) || files.length === 0) {
                return;
            }
            setSelectedFiles((prev) => mergeUniquePaths(prev, files));
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイル選択に失敗しました'));
        }
    }, []);

    const handleRemoveFile = useCallback((pathToRemove: string) => {
        setSelectedFiles((prev) => prev.filter((path) => path !== pathToRemove));
    }, []);

    const handleLoadSelectedFiles = useCallback(async () => {
        if (selectedFiles.length === 0) {
            return;
        }

        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await LoadTranslationFlowFiles(taskId, selectedFiles);
            const mapped = mapLoadResult(payload);
            if (mapped.taskId !== '' && mapped.taskId !== taskId) {
                setTaskID(mapped.taskId);
            }
            setLoadedFiles(mapped.files);
            setSelectedFiles([]);
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイルロードに失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, [selectedFiles, taskId]);

    const handlePreviewPageChange = useCallback(async (fileId: number, page: number) => {
        if (fileId <= 0) {
            return;
        }

        const safePage = Math.max(1, page);
        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await ListTranslationFlowPreviewRows(fileId, safePage, PREVIEW_PAGE_SIZE);
            const mappedPage = mapPreviewPage(payload);
            setLoadedFiles((prev) =>
                prev.map((file) => {
                    if (file.fileId !== fileId) {
                        return file;
                    }
                    return {
                        ...file,
                        currentPage: mappedPage.page,
                        pageSize: mappedPage.pageSize,
                        totalRows: mappedPage.totalRows,
                        rows: mappedPage.rows,
                    };
                }),
            );
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'プレビューのページ切り替えに失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, []);

    const handleAdvanceFromLoad = useCallback(() => {
        if (loadedFiles.length === 0) {
            return;
        }
        setActiveTab(1);
    }, [loadedFiles.length]);

    const handleTabChange = useCallback((index: number) => {
        if (index < 0 || index >= TABS.length) {
            return;
        }
        setActiveTab(index);
    }, []);

    return {
        state: {
            taskId,
            activeTab,
            tabs: TABS,
            selectedFiles,
            loadedFiles,
            isLoading,
            errorMessage,
        },
        actions: {
            handleTabChange,
            handleSelectFiles,
            handleRemoveFile,
            handleLoadSelectedFiles,
            handleReloadFiles,
            handlePreviewPageChange,
            handleAdvanceFromLoad,
        },
        ui: {
            previewPageSize: PREVIEW_PAGE_SIZE,
        },
    };
}
