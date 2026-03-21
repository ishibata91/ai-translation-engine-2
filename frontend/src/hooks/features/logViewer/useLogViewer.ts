import {type MouseEvent as ReactMouseEvent, useCallback, useEffect, useRef, useState} from 'react';
import {z} from 'zod';
import type {LogEntry} from '../../../components/log/LogDetail';
import {useUIStore} from '../../../store/uiStore';
import {UIStateGetJSON, UIStateSetJSON} from '../../../wailsjs/go/controller/ConfigController';
import {useLogStore} from './useLogStore';

const UI_NAMESPACE = 'log-viewer';
const FILTER_KEY = 'filters';

const logLevelSchema = z.enum(['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR']);
const savedFiltersSchema = z.object({
    level: logLevelSchema.optional(),
    traceId: z.string().optional(),
});

type LogLevel = z.infer<typeof logLevelSchema>;

interface LogFilters {
    level: LogLevel;
    traceId: string;
}

const DEFAULT_FILTERS: LogFilters = {
    level: 'ERROR',
    traceId: '',
};

const LEVEL_ORDER: Record<LogEntry['level'], number> = {
    DEBUG: 0,
    INFO: 1,
    WARN: 2,
    ERROR: 3,
};

function matchesFilter(log: LogEntry, filters: LogFilters): boolean {
    if (filters.level !== 'ALL') {
        const logLevelNum = LEVEL_ORDER[log.level];
        const filterLevelNum = LEVEL_ORDER[filters.level];
        if (logLevelNum < filterLevelNum) {
            return false;
        }
    }

    if (filters.traceId.trim()) {
        const traceID = log.attributes?.trace_id ?? '';
        if (!String(traceID).toLowerCase().includes(filters.traceId.trim().toLowerCase())) {
            return false;
        }
    }

    return true;
}

/**
 * LogViewer の永続化、イベント購読、表示用 state を集約する。
 * ログの蓄積・定期掃除は useLogStore に委譲する。
 */
export function useLogViewer() {
    const { logViewerWidth: width, setLogViewerWidth: setWidth, setDetailPane } = useUIStore();
    const [isCollapsed, setIsCollapsed] = useState(true);
    const [isResizing, setIsResizing] = useState(false);
    const [filters, setFilters] = useState<LogFilters>(DEFAULT_FILTERS);
    const filtersLoaded = useRef(false);

    // ログ蓄積・定期掃除は useLogStore に委譲
    const { allLogs, allLogsCount } = useLogStore();

    const filteredLogs = allLogs.filter((log) => matchesFilter(log, filters));

    const handleMouseDown = useCallback((event: ReactMouseEvent) => {
        event.preventDefault();
        setIsResizing(true);
    }, []);

    useEffect(() => {
        if (!isResizing) {
            return;
        }

        const handleMouseMove = (event: MouseEvent) => {
            let nextWidth = window.innerWidth - event.clientX;
            nextWidth = Math.max(200, Math.min(nextWidth, 800));
            setWidth(nextWidth);
        };
        const handleMouseUp = () => setIsResizing(false);

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = 'none';

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.userSelect = '';
        };
    }, [isResizing, setWidth]);

    useEffect(() => {
        void UIStateGetJSON(UI_NAMESPACE, FILTER_KEY)
            .then((jsonStr) => {
                if (!jsonStr) {
                    return;
                }

                const parsed: unknown = JSON.parse(jsonStr);
                const result = savedFiltersSchema.safeParse(parsed);
                if (!result.success) {
                    return;
                }

                setFilters((prev) => ({
                    ...prev,
                    level: result.data.level ?? prev.level,
                    traceId: result.data.traceId ?? prev.traceId,
                }));
            })
            .catch(() => {
                // 永続化データが壊れていても既定値で続行する。
            })
            .finally(() => {
                filtersLoaded.current = true;
            });
    }, []);

    useEffect(() => {
        if (!filtersLoaded.current) {
            return;
        }

        const payload: LogFilters = {
            level: filters.level,
            traceId: filters.traceId,
        };
        void UIStateSetJSON(UI_NAMESPACE, FILTER_KEY, payload).catch(() => {
            // 保存失敗時も表示は継続する。
        });
    }, [filters]);

    const handleLogClick = (log: LogEntry) => {
        setDetailPane(true, 'log', log);
    };

    const handleLevelChange = (rawLevel: string) => {
        const parsed = logLevelSchema.safeParse(rawLevel);
        if (!parsed.success) {
            return;
        }
        setFilters((prev) => ({ ...prev, level: parsed.data }));
    };

    const handleTraceIDChange = (traceId: string) => {
        setFilters((prev) => ({ ...prev, traceId }));
    };

    return {
        allLogsCount,
        filteredLogs,
        filters,
        handleLevelChange,
        handleLogClick,
        handleMouseDown,
        handleTraceIDChange,
        isCollapsed,
        setIsCollapsed,
        width,
    };
}
