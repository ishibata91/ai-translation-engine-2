export type DebuggerLogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error';

export interface DebuggerLogEvent {
    timestamp: string;
    level: DebuggerLogLevel;
    scope: string;
    stage: string;
    message: string;
    fields?: Record<string, unknown>;
}

declare global {
    interface Window {
        __AITE2_DEBUGGER_LOGS__?: DebuggerLogEvent[];
    }
}

const debuggerBuffer = (): DebuggerLogEvent[] => {
    if (typeof window === 'undefined') {
        return [];
    }
    window.__AITE2_DEBUGGER_LOGS__ ??= [];
    return window.__AITE2_DEBUGGER_LOGS__;
};

// createDebuggerLogger is intentionally isolated under frontend/src/debugger so
// import lines and call sites can be removed in one sweep after a bugfix run.
export const createDebuggerLogger = (scope: string) => {
    const write = (level: DebuggerLogLevel, stage: string, message: string, fields?: Record<string, unknown>) => {
        const event: DebuggerLogEvent = {
            timestamp: new Date().toISOString(),
            level,
            scope,
            stage,
            message,
            fields,
        };
        debuggerBuffer().push(event);
        const consoleMethod = level === 'trace' || level === 'debug' ? 'debug' : level;
        console[consoleMethod](`[debugger:${scope}] ${stage} ${message}`, fields ?? {});
        return event;
    };

    return {
        trace: (stage: string, message: string, fields?: Record<string, unknown>) => write('trace', stage, message, fields),
        debug: (stage: string, message: string, fields?: Record<string, unknown>) => write('debug', stage, message, fields),
        info: (stage: string, message: string, fields?: Record<string, unknown>) => write('info', stage, message, fields),
        warn: (stage: string, message: string, fields?: Record<string, unknown>) => write('warn', stage, message, fields),
        error: (stage: string, message: string, fields?: Record<string, unknown>) => write('error', stage, message, fields),
    };
};

export const readDebuggerLogs = (): DebuggerLogEvent[] => [...debuggerBuffer()];

export const clearDebuggerLogs = (): void => {
    if (typeof window === 'undefined' || window.__AITE2_DEBUGGER_LOGS__ === undefined) {
        return;
    }
    window.__AITE2_DEBUGGER_LOGS__ = [];
};
