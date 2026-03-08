import React, { useState, useEffect, useCallback, useRef } from 'react';
import type { LogEntry } from './LogDetail';
import { useUIStore } from '../../store/uiStore';
import { useWailsEvent } from '../../hooks/useWailsEvent';
import { UIStateGetJSON, UIStateSetJSON } from '../../wailsjs/go/config/ConfigService';

const WAILS_LOG_EVENT = 'telemetry.log';
const UI_NAMESPACE = 'log-viewer';
const FILTER_KEY = 'filters';
const MAX_LOG_ENTRIES = 500;

type LogLevel = 'ALL' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

interface LogFilters {
    level: LogLevel;
    traceId: string;
}

const DEFAULT_FILTERS: LogFilters = {
    level: 'ERROR',
    traceId: '',
};

const LEVEL_ORDER: Record<string, number> = {
    DEBUG: 0,
    INFO: 1,
    WARN: 2,
    ERROR: 3,
};

let eventCounter = 0;

function matchesFilter(log: LogEntry, filters: LogFilters): boolean {
    if (filters.level !== 'ALL') {
        const logLevelNum = LEVEL_ORDER[log.level] ?? -1;
        const filterLevelNum = LEVEL_ORDER[filters.level] ?? -1;
        if (logLevelNum < filterLevelNum) return false;
    }

    if (filters.traceId.trim()) {
        const tid = log.attributes?.['trace_id'] ?? '';
        if (!String(tid).toLowerCase().includes(filters.traceId.trim().toLowerCase())) return false;
    }

    return true;
}

const LogItem: React.FC<{ log: LogEntry; onClick: (log: LogEntry) => void }> = ({ log, onClick }) => {
    const [isNew, setIsNew] = useState(true);
    const traceId = log.attributes?.['trace_id'];

    useEffect(() => {
        const timer = setTimeout(() => setIsNew(false), 500);
        return () => clearTimeout(timer);
    }, []);

    const baseBg =
        log.level === 'ERROR'
            ? 'alert-error'
            : log.level === 'WARN'
                ? 'alert-warning'
                : log.level === 'INFO'
                    ? 'bg-base-100'
                    : 'bg-base-200 text-base-content/70';

    return (
        <div
            className={`alert p-2 rounded flex-col items-start gap-1 cursor-pointer transition-all duration-500 ease-out hover:opacity-80 ${
                isNew ? 'bg-secondary text-secondary-content scale-[1.02] shadow-lg shadow-secondary/20 z-10' : baseBg
            }`}
            onClick={() => onClick(log)}
        >
            <span className="font-semibold mix-blend-hard-light">[{log.level}] {log.message}</span>
            {Boolean(traceId) && (
                <span className={`badge badge-sm font-mono opacity-70 text-[0.6rem] ${isNew ? 'border border-secondary-content/20 bg-transparent text-secondary-content' : 'badge-ghost'}`}>
                    TraceID: {String(traceId).slice(0, 8)}
                </span>
            )}
        </div>
    );
};

const LogViewer: React.FC = () => {
    const { logViewerWidth: width, setLogViewerWidth: setWidth, setDetailPane } = useUIStore();
    const [isResizing, setIsResizing] = useState(false);
    const [isCollapsed, setIsCollapsed] = useState(true);
    const [allLogs, setAllLogs] = useState<LogEntry[]>([]);
    const [filters, setFilters] = useState<LogFilters>(DEFAULT_FILTERS);
    const filtersLoaded = useRef(false);

    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        e.preventDefault();
        setIsResizing(true);
    }, []);

    useEffect(() => {
        if (!isResizing) return;
        const handleMouseMove = (e: MouseEvent) => {
            let newWidth = window.innerWidth - e.clientX;
            newWidth = Math.max(200, Math.min(newWidth, 800));
            setWidth(newWidth);
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
        UIStateGetJSON(UI_NAMESPACE, FILTER_KEY)
            .then((jsonStr) => {
                if (!jsonStr) return;
                try {
                    const parsed = JSON.parse(jsonStr) as Partial<LogFilters>;
                    setFilters((prev) => ({
                        ...prev,
                        level: (parsed.level as LogLevel) ?? 'ERROR',
                        traceId: parsed.traceId ?? '',
                    }));
                } catch {
                    // noop
                }
            })
            .catch(() => {
                // noop
            })
            .finally(() => {
                filtersLoaded.current = true;
            });
    }, []);

    useEffect(() => {
        if (!filtersLoaded.current) return;
        const payload = { level: filters.level, traceId: filters.traceId };
        UIStateSetJSON(UI_NAMESPACE, FILTER_KEY, payload).catch(() => {
            // noop
        });
    }, [filters]);

    useWailsEvent<unknown>(WAILS_LOG_EVENT, (event) => {
        if (!event || typeof event !== 'object') return;
        const raw = event as Record<string, unknown>;
        const levelRaw = raw['level'];
        const normalizedLevel: LogEntry['level'] =
            levelRaw === 'DEBUG' || levelRaw === 'INFO' || levelRaw === 'WARN' || levelRaw === 'ERROR'
                ? levelRaw
                : 'INFO';

        const entry: LogEntry = {
            id: `${raw['id'] ?? Date.now()}-${eventCounter++}`,
            level: normalizedLevel,
            message: String(raw['message'] ?? ''),
            timestamp: String(raw['timestamp'] ?? new Date().toISOString()),
            attributes: (raw['attributes'] as Record<string, unknown>) ?? {},
        };

        setAllLogs((prev) => {
            const next = [entry, ...prev];
            return next.length > MAX_LOG_ENTRIES ? next.slice(0, MAX_LOG_ENTRIES) : next;
        });
    });

    const filteredLogs = allLogs.filter((log) => matchesFilter(log, filters));

    const handleLogClick = (log: LogEntry) => {
        setDetailPane(true, 'log', log);
    };

    const handleLevelChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        setFilters((prev) => ({ ...prev, level: e.target.value as LogLevel }));
    };

    const handleTraceIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFilters((prev) => ({ ...prev, traceId: e.target.value }));
    };

    if (isCollapsed) {
        return (
            <div className="h-full bg-base-300 border-l border-base-200 p-2 flex flex-col items-center">
                <button
                    className="btn btn-ghost btn-sm btn-square"
                    onClick={() => setIsCollapsed(false)}
                    title="Open Logs"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" /></svg>
                </button>
                <div className="mt-4 text-xs font-bold text-base-content/50 tracking-widest hidden sm:block" style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}>
                    LOGS
                </div>
            </div>
        );
    }

    return (
        <div
            style={{ width: `${width}px` }}
            className="group relative flex flex-col p-4 gap-4 h-full bg-base-300 border-l border-base-200 shadow-md transition-shadow"
        >
            <div
                className="absolute left-0 top-0 bottom-0 w-1.5 cursor-col-resize hover:bg-primary z-20 transition-colors"
                onMouseDown={handleMouseDown}
            />

            <div className="flex items-center justify-between gap-2 text-sm z-10 relative">
                <div className="flex items-center gap-2">
                    <div className="badge badge-outline">Telemetry Logs</div>
                    {allLogs.length > 0 && (
                        <span className="text-xs opacity-40">
                            {filteredLogs.length}/{allLogs.length}
                        </span>
                    )}
                </div>
                <button
                    className="btn btn-ghost btn-xs btn-square shrink-0"
                    onClick={() => setIsCollapsed(true)}
                    title="Collapse Logs"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" /></svg>
                </button>
            </div>

            <div className="flex flex-col gap-2 w-full z-10 relative">
                <select
                    className="select select-bordered select-sm w-full font-sans"
                    value={filters.level}
                    onChange={handleLevelChange}
                    id="log-level-filter"
                >
                    <option value="ALL">All Levels</option>
                    <option value="INFO">INFO以上</option>
                    <option value="WARN">WARN以上</option>
                    <option value="ERROR">ERROR</option>
                    <option value="DEBUG">DEBUG以上</option>
                </select>
                <input
                    id="log-traceid-filter"
                    type="text"
                    placeholder="Filter by TraceID..."
                    className="input input-bordered input-sm w-full font-mono text-xs"
                    value={filters.traceId}
                    onChange={handleTraceIdChange}
                />
            </div>

            <div className="flex flex-col gap-2 overflow-y-auto w-full h-full text-xs z-0 relative pr-1">
                {filteredLogs.length === 0 && (
                    <div className="text-center opacity-30 mt-8 text-xs">
                        {allLogs.length === 0 ? 'ログ受信待ち...' : 'フィルター条件に一致するログなし'}
                    </div>
                )}
                {filteredLogs.map((log) => (
                    <LogItem key={log.id} log={log} onClick={handleLogClick} />
                ))}
            </div>
        </div>
    );
};

export default LogViewer;
