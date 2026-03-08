import React, { useState, useEffect } from 'react';
import type { LogEntry } from './LogDetail';
import { useLogViewer } from '../../hooks/features/logViewer/useLogViewer';

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
    const {
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
    } = useLogViewer();

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
                    {allLogsCount > 0 && (
                        <span className="text-xs opacity-40">
                            {filteredLogs.length}/{allLogsCount}
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
                    onChange={(e) => handleLevelChange(e.target.value)}
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
                    onChange={(e) => handleTraceIDChange(e.target.value)}
                />
            </div>

            <div className="flex flex-col gap-2 overflow-y-auto w-full h-full text-xs z-0 relative pr-1">
                {filteredLogs.length === 0 && (
                    <div className="text-center opacity-30 mt-8 text-xs">
                        {allLogsCount === 0 ? 'ログ受信待ち...' : 'フィルター条件に一致するログなし'}
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
