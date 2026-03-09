import { useCallback, useEffect, useRef, useState, type MouseEvent as ReactMouseEvent } from 'react';
import { z } from 'zod';
import type { LogEntry } from '../../../components/log/LogDetail';
import { useUIStore } from '../../../store/uiStore';
import { UIStateGetJSON, UIStateSetJSON } from '../../../wailsjs/go/controller/ConfigController';
import { useWailsEvent } from '../../useWailsEvent';

const WAILS_LOG_EVENT = 'telemetry.log';
const UI_NAMESPACE = 'log-viewer';
const FILTER_KEY = 'filters';
const MAX_LOG_ENTRIES = 500;

const logLevelSchema = z.enum(['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR']);
const savedFiltersSchema = z.object({
    level: logLevelSchema.optional(),
    traceId: z.string().optional(),
});
const logEventSchema = z.object({
    id: z.union([z.string(), z.number()]).optional(),
    level: z.enum(['DEBUG', 'INFO', 'WARN', 'ERROR']).optional(),
    message: z.string().optional(),
    timestamp: z.string().optional(),
    attributes: z.record(z.string(), z.unknown()).optional(),
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

let eventCounter = 0;

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
 */
export function useLogViewer() {
    const { logViewerWidth: width, setLogViewerWidth: setWidth, setDetailPane } = useUIStore();
    const [isCollapsed, setIsCollapsed] = useState(true);
    const [isResizing, setIsResizing] = useState(false);
    const [allLogs, setAllLogs] = useState<LogEntry[]>([]);
    const [filters, setFilters] = useState<LogFilters>(DEFAULT_FILTERS);
    const filtersLoaded = useRef(false);

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

    useWailsEvent<unknown>(WAILS_LOG_EVENT, (payload) => {
        const parsed = logEventSchema.safeParse(payload);
        if (!parsed.success) {
            return;
        }

        const entry: LogEntry = {
            id: `${parsed.data.id ?? Date.now()}-${eventCounter++}`,
            level: parsed.data.level ?? 'INFO',
            message: parsed.data.message ?? '',
            timestamp: parsed.data.timestamp ?? new Date().toISOString(),
            attributes: parsed.data.attributes ?? {},
        };

        setAllLogs((prev) => {
            const next = [entry, ...prev];
            return next.length > MAX_LOG_ENTRIES ? next.slice(0, MAX_LOG_ENTRIES) : next;
        });
    });

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
        allLogsCount: allLogs.length,
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
