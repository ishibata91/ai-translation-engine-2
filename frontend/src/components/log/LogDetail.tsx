import React from 'react';

// バックエンドのWailsハンドラーが送信するログイベントの型定義
// WailsLogEvent (Go側) と対応している
export interface LogEntry {
    id: string;
    level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG';
    message: string;
    timestamp: string;
    // バックエンドの全てのテレメトリ属性を動的に受け取るフィールド
    // trace_id, action, resource_type, resource_id, env, app_version, service_name, host_name などが含まれる
    attributes: Record<string, any>;
}

interface LogDetailProps {
    log: LogEntry | null;
}

// 既知テレメトリ属性（上部に専用UIで表示するもの）
const KNOWN_CONTEXT_KEYS = new Set(['trace_id', 'action', 'resource_type', 'resource_id']);
// システムグローバル属性（折り畳んで表示するもの）
const SYSTEM_KEYS = new Set(['env', 'app_version', 'service_name', 'host_name']);

const LogDetail: React.FC<LogDetailProps> = ({ log }) => {
    if (!log) return null;

    const { attributes } = log;
    const traceId = attributes?.['trace_id'];
    const action = attributes?.['action'];
    const resourceType = attributes?.['resource_type'];
    const resourceId = attributes?.['resource_id'];
    const stackTrace = attributes?.['stack_trace'] ?? attributes?.['error.stack'];

    // 「その他」属性: 既知/システム属性以外
    const extraEntries = Object.entries(attributes ?? {}).filter(
        ([k]) => !KNOWN_CONTEXT_KEYS.has(k) && !SYSTEM_KEYS.has(k) && k !== 'stack_trace' && k !== 'error.stack'
    );
    const systemEntries = Object.entries(attributes ?? {}).filter(([k]) => SYSTEM_KEYS.has(k));

    return (
        <div className="flex flex-col gap-4 p-2 h-full">
            {/* ヘッダー */}
            <div className="flex items-center gap-4 border-b border-base-300 pb-2 shrink-0">
                <div className={`badge ${log.level === 'ERROR' ? 'badge-error' : log.level === 'WARN' ? 'badge-warning' : log.level === 'INFO' ? 'badge-info' : 'badge-ghost'}`}>
                    {log.level}
                </div>
                <div className="text-lg font-bold truncate">{log.message}</div>
                <div className="ml-auto text-xs opacity-50 shrink-0">{log.timestamp}</div>
            </div>

            <div className="grid grid-cols-2 gap-4 flex-1 min-h-0">
                {/* Left Column: Trace Context & Action */}
                <div className="flex flex-col gap-4 overflow-y-auto pr-2">
                    <div>
                        <h4 className="font-bold text-sm mb-2 text-primary">Trace Context</h4>
                        <div className="bg-base-200 p-3 rounded-md font-mono text-xs flex flex-col gap-1">
                            {traceId && <div><span className="opacity-50">trace_id:</span> {traceId}</div>}
                            {action && <div><span className="opacity-50">action:</span> {action}</div>}
                            {resourceType && <div><span className="opacity-50">resource_type:</span> {resourceType}</div>}
                            {resourceId && <div><span className="opacity-50">resource_id:</span> {resourceId}</div>}
                            {!traceId && !action && !resourceType && (
                                <div className="opacity-40 italic">コンテキスト情報なし</div>
                            )}
                        </div>
                    </div>

                    {stackTrace && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-error">Stack Trace</h4>
                            <pre className="bg-base-200 p-3 rounded-md font-mono text-xs overflow-x-auto text-error whitespace-pre-wrap">
                                {String(stackTrace)}
                            </pre>
                        </div>
                    )}
                </div>

                {/* Right Column: Extra Attributes & System Info */}
                <div className="flex flex-col gap-4 overflow-y-auto pr-2">
                    {extraEntries.length > 0 && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-secondary">Attributes</h4>
                            <div className="bg-base-200 p-3 rounded-md font-mono text-xs flex flex-col gap-1">
                                {extraEntries.map(([k, v]) => (
                                    <div key={k}>
                                        <span className="opacity-50">{k}:</span>{' '}
                                        {typeof v === 'object' ? JSON.stringify(v) : String(v)}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {systemEntries.length > 0 && (
                        <div>
                            <h4 className="font-bold text-sm mb-2 text-accent opacity-60">System</h4>
                            <div className="bg-base-200 p-3 rounded-md font-mono text-xs flex flex-col gap-1 opacity-60">
                                {systemEntries.map(([k, v]) => (
                                    <div key={k}>
                                        <span className="opacity-50">{k}:</span> {String(v)}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default LogDetail;
