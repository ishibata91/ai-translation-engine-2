import React from 'react';

export interface LogEntry {
    id: string;
    level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG';
    message: string;
    traceId: string;
    spanId?: string;
    timestamp: string;
    requestPayload?: string;
    responsePayload?: string;
    stackTrace?: string;
}

interface LogDetailProps {
    log: LogEntry | null;
}

const LogDetail: React.FC<LogDetailProps> = ({ log }) => {
    if (!log) return null;

    return (
        <div className="flex flex-col gap-4 p-2 h-full">
            <div className="flex items-center gap-4 border-b border-base-300 pb-2 shrink-0">
                <div className={`badge ${log.level === 'ERROR' ? 'badge-error' : log.level === 'WARN' ? 'badge-warning' : 'badge-info'}`}>
                    {log.level}
                </div>
                <div className="text-lg font-bold">{log.message}</div>
                <div className="ml-auto text-xs opacity-50">{log.timestamp}</div>
            </div>

            <div className="grid grid-cols-2 gap-4 flex-1 min-h-0">
                {/* Left Column: Trace Context & Basics */}
                <div className="flex flex-col gap-4 overflow-y-auto pr-2">
                    <div>
                        <h4 className="font-bold text-sm mb-2 text-primary">Trace Context</h4>
                        <div className="bg-base-200 p-3 rounded-md font-mono text-xs flex flex-col gap-1">
                            <div><span className="opacity-50">TraceID:</span> {log.traceId}</div>
                            {log.spanId && <div><span className="opacity-50">SpanID:</span> {log.spanId}</div>}
                        </div>
                    </div>

                    {log.stackTrace && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-error">Stack Trace</h4>
                            <pre className="bg-base-200 p-3 rounded-md font-mono text-xs overflow-x-auto text-error whitespace-pre-wrap">
                                {log.stackTrace}
                            </pre>
                        </div>
                    )}
                </div>

                {/* Right Column: Request / Response Payload */}
                <div className="flex flex-col gap-4 overflow-y-auto pr-2">
                    {log.requestPayload && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-secondary">Request Payload</h4>
                            <pre className="bg-base-200 p-3 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap">
                                {log.requestPayload}
                            </pre>
                        </div>
                    )}

                    {log.responsePayload && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-accent">Response Payload</h4>
                            <pre className="bg-base-200 p-3 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap">
                                {log.responsePayload}
                            </pre>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default LogDetail;
