import React, { useState, useEffect, useCallback } from 'react';
import type { LogEntry } from './LogDetail';

interface LogViewerProps {
    onLogClick?: (log: LogEntry) => void;
}

const MOCK_LOGS: LogEntry[] = [
    {
        id: 'log-1',
        level: 'INFO',
        message: 'Parsing mod structure...',
        traceId: '123e4567-e89b-12d3',
        spanId: 'span-a1b2',
        timestamp: '2026-02-28T14:45:01',
    },
    {
        id: 'log-2',
        level: 'DEBUG',
        message: 'Loaded terminology dictionary',
        traceId: '123e4567-e89b-12d3',
        spanId: 'span-c3d4',
        timestamp: '2026-02-28T14:45:02',
        requestPayload: '{\n  "path": "/data/dictionary.json",\n  "useCache": true\n}',
        responsePayload: '{\n  "status": "success",\n  "entriesCount": 12450\n}'
    },
    {
        id: 'log-3',
        level: 'ERROR',
        message: 'Rate limit exceeded during translation.',
        traceId: '987f6543-a21b-45c6',
        spanId: 'span-e5f6',
        timestamp: '2026-02-28T14:45:05',
        requestPayload: '{\n  "model": "gemini-2.0-flash",\n  "prompt": "Translate the following..."\n}',
        stackTrace: 'Error: Rate limit exceeded (429)\n  at Translator.callLLM (/src/translation/translator.ts:55:15)\n  at async JobRunner.processBatch (/src/jobs/runner.ts:112:21)',
    }
];

const LogViewer: React.FC<LogViewerProps> = ({ onLogClick }) => {
    const [width, setWidth] = useState(320);
    const [isResizing, setIsResizing] = useState(false);

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
    }, [isResizing]);

    return (
        <div
            style={{ width: `${width}px` }}
            className="group relative flex flex-col p-4 gap-4 h-full bg-base-300 border-l border-base-200 shadow-md transition-shadow"
        >
            <div
                className="absolute left-0 top-0 bottom-0 w-1.5 cursor-col-resize hover:bg-primary z-20 transition-colors"
                onMouseDown={handleMouseDown}
            />

            <div className="flex items-center text-sm">
                <div className="badge badge-outline">Telemetry Logs</div>
            </div>

            <div className="flex flex-col gap-2 w-full">
                <select className="select select-bordered select-sm w-full">
                    <option value="">All Levels</option>
                    <option value="INFO">INFO</option>
                    <option value="WARN">WARN</option>
                    <option value="ERROR">ERROR</option>
                    <option value="DEBUG">DEBUG</option>
                </select>
                <input
                    type="text"
                    placeholder="Filter by TraceID..."
                    className="input input-bordered input-sm w-full font-mono text-xs"
                />
            </div>

            <div className="flex flex-col gap-2 overflow-y-auto w-full h-full text-xs">
                {MOCK_LOGS.map((log) => (
                    <div
                        key={log.id}
                        className={`alert p-2 rounded flex-col items-start gap-1 cursor-pointer hover:opacity-80 transition-opacity ${log.level === 'ERROR' ? 'alert-error' :
                            log.level === 'WARN' ? 'alert-warning' :
                                log.level === 'INFO' ? 'bg-base-100' : 'bg-base-200 text-base-content/70'
                            }`}
                        onClick={() => onLogClick?.(log)}
                    >
                        <span>[{log.level}] {log.message}</span>
                        <span className="badge badge-ghost badge-sm font-mono opacity-50 text-[0.6rem]">TraceID: {log.traceId.slice(0, 8)}</span>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default LogViewer;
