import React, { useState, useEffect, useCallback, useRef } from 'react';
import type { LogEntry } from './LogDetail';
import { useUIStore } from '../../store/uiStore';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { UIStateGetJSON, UIStateSetJSON } from '../../wailsjs/go/config/ConfigService';

// Wails バックエンドから届くイベント名（Go側: wailsLogEventName と一致させる）
const WAILS_LOG_EVENT = 'telemetry.log';

// ConfigService で使用する永続化キー
const UI_NAMESPACE = 'log-viewer';
const FILTER_KEY = 'filters';

// 一度に保持するログの最大数（パフォーマンス保護）
const MAX_LOG_ENTRIES = 500;

type LogLevel = 'ALL' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

interface LogFilters {
    level: LogLevel;
    traceId: string;
}

const DEFAULT_FILTERS: LogFilters = {
    level: 'ERROR', // 仕様: デフォルトは ERROR
    traceId: '',
};

const LEVEL_ORDER: Record<string, number> = {
    DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3,
};

// React の key 重複によるコンポーネント再利用バグを防ぐための一意カウンタ
let eventCounter = 0;

function matchesFilter(log: LogEntry, filters: LogFilters): boolean {
    // レベルフィルター
    if (filters.level !== 'ALL') {
        const logLevelNum = LEVEL_ORDER[log.level] ?? -1;
        const filterLevelNum = LEVEL_ORDER[filters.level] ?? -1;
        if (logLevelNum < filterLevelNum) return false;
    }
    // TraceID フィルター
    if (filters.traceId.trim()) {
        const tid = log.attributes?.['trace_id'] ?? '';
        if (!String(tid).toLowerCase().includes(filters.traceId.trim().toLowerCase())) return false;
    }
    return true;
}

const LogItem: React.FC<{ log: LogEntry; onClick: (log: LogEntry) => void }> = ({ log, onClick }) => {
    // マウント時（新しくリストに追加されたとき）に0.5秒間 true になるフラグ
    const [isNew, setIsNew] = useState(true);
    const traceId = log.attributes?.['trace_id'];

    useEffect(() => {
        const timer = setTimeout(() => setIsNew(false), 500);
        return () => clearTimeout(timer);
    }, []);

    const baseBg = log.level === 'ERROR' ? 'alert-error' :
        log.level === 'WARN' ? 'alert-warning' :
            log.level === 'INFO' ? 'bg-base-100' : 'bg-base-200 text-base-content/70';

    return (
        <div
            className={`alert p-2 rounded flex-col items-start gap-1 cursor-pointer transition-all duration-500 ease-out hover:opacity-80 ${isNew ? 'bg-secondary text-secondary-content scale-[1.02] shadow-lg shadow-secondary/20 z-10' : baseBg
                }`}
            onClick={() => onClick(log)}
        >
            <span className="font-semibold mix-blend-hard-light">[{log.level}] {log.message}</span>
            {traceId && (
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

    // ログ一覧（全件）
    const [allLogs, setAllLogs] = useState<LogEntry[]>([]);
    // フィルター設定
    const [filters, setFilters] = useState<LogFilters>(DEFAULT_FILTERS);
    // 設定読込済みフラグ
    const filtersLoaded = useRef(false);

    // --- リサイズ処理 ---
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

    // --- タスク 7.1, 7.2: マウント時に保存済みフィルター設定を読み込む ---
    useEffect(() => {
        UIStateGetJSON(UI_NAMESPACE, FILTER_KEY)
            .then((jsonStr) => {
                if (jsonStr) {
                    try {
                        const parsed = JSON.parse(jsonStr) as Partial<LogFilters>;
                        setFilters((prev) => ({
                            ...prev,
                            level: (parsed.level as LogLevel) ?? 'ERROR',
                            traceId: parsed.traceId ?? '',
                        }));
                    } catch {
                        // パース失敗の場合はデフォルトのまま
                    }
                }
            })
            .catch(() => { })
            .finally(() => {
                filtersLoaded.current = true;
            });
    }, []);

    // --- タスク 7.3: フィルター変更時に即座に永続化 ---
    useEffect(() => {
        if (!filtersLoaded.current) return;
        const payload = { level: filters.level, traceId: filters.traceId };
        UIStateSetJSON(UI_NAMESPACE, FILTER_KEY, payload).catch(() => { });
    }, [filters]);

    // --- タスク 5.1, 5.2: Wails イベントでログを受信 ---
    useEffect(() => {
        const handleLog = (event: unknown) => {
            if (!event || typeof event !== 'object') return;
            const raw = event as Record<string, any>;
            const entry: LogEntry = {
                id: `${raw['id'] ?? Date.now()}-${eventCounter++}`,
                level: (['DEBUG', 'INFO', 'WARN', 'ERROR'].includes(raw['level'])
                    ? raw['level']
                    : 'INFO') as LogEntry['level'],
                message: String(raw['message'] ?? ''),
                timestamp: String(raw['timestamp'] ?? new Date().toISOString()),
                attributes: (raw['attributes'] as Record<string, any>) ?? {},
            };

            setAllLogs((prev) => {
                const next = [entry, ...prev];
                return next.length > MAX_LOG_ENTRIES ? next.slice(0, MAX_LOG_ENTRIES) : next;
            });
        };

        EventsOn(WAILS_LOG_EVENT, handleLog);
        return () => {
            EventsOff(WAILS_LOG_EVENT);
        };
    }, []);

    // --- タスク 5.3: フィルター適用 ---
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
            {/* リサイズハンドル */}
            <div
                className="absolute left-0 top-0 bottom-0 w-1.5 cursor-col-resize hover:bg-primary z-20 transition-colors"
                onMouseDown={handleMouseDown}
            />

            {/* タイトル＆ログ件数 */}
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

            {/* フィルターコントロール */}
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

            {/* ログ一覧 */}
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
