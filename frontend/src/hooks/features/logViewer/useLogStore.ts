/**
 * useLogStore — インメモリ LogEntry リングバッファ + 定期掃除ロガー
 *
 * 責務:
 *  - Wails telemetry.log イベントを受信して LogEntry に変換・蓄積する
 *  - MAX_LOG_ENTRIES を超えた分は先着順で削除（リングバッファ）
 *  - LOG_EXPIRE_MS を超えた古いエントリを CLEANUP_INTERVAL_MS ごとに自動削除
 *
 * 利用方法:
 *  const { allLogs, allLogsCount } = useLogStore();
 */

import {useCallback, useEffect, useRef, useState} from 'react';
import {z} from 'zod';
import type {LogEntry} from '../../../components/log/LogDetail';
import {useWailsEvent} from '../../useWailsEvent';

// ---- 定数 -------------------------------------------------------------------

/** Wails イベント名（バックエンド telemetry パッケージと一致させること） */
const WAILS_LOG_EVENT = 'telemetry.log';

/** インメモリに保持するログの最大件数 */
const MAX_LOG_ENTRIES = 500;

/** ログエントリの有効期限（ミリ秒）: 30 分 */
const LOG_EXPIRE_MS = 30 * 60 * 1000;

/** 定期掃除の間隔（ミリ秒）: 5 分 */
const CLEANUP_INTERVAL_MS = 5 * 60 * 1000;

// ---- スキーマ ----------------------------------------------------------------

const logEventSchema = z.object({
  id: z.union([z.string(), z.number()]).optional(),
  level: z.enum(['DEBUG', 'INFO', 'WARN', 'ERROR']).optional(),
  message: z.string().optional(),
  timestamp: z.string().optional(),
  attributes: z.record(z.string(), z.unknown()).optional(),
});

// ---- カウンタ（モジュールスコープ）------------------------------------------

let eventCounter = 0;

// ---- Helpers ----------------------------------------------------------------

/**
 * ログエントリの receivedAt を付与した拡張型。
 * useLogStore 内でのみ使用し、外部には LogEntry として公開する。
 */
type LogEntryWithMeta = LogEntry & { receivedAt: number };

/**
 * ISO 8601 タイムスタンプ文字列をミリ秒エポックに変換する。
 * パース失敗時は現在時刻を返す。
 */
function parseTimestampMs(ts: string | undefined): number {
  if (!ts) return Date.now();
  const ms = Date.parse(ts);
  return Number.isNaN(ms) ? Date.now() : ms;
}

// ---- Hook -------------------------------------------------------------------

/**
 * useLogStore は Wails telemetry.log イベントを購読し、
 * ログエントリをインメモリに保持する。
 *
 * 定期的に LOG_EXPIRE_MS を超えた古いエントリを削除する。
 */
export function useLogStore() {
  const [allLogs, setAllLogs] = useState<LogEntryWithMeta[]>([]);
  const cleanupTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // ---- 定期掃除（マウント時に起動、アンマウント時にクリア）-----------------
  useEffect(() => {
    const cleanup = () => {
      const expireThreshold = Date.now() - LOG_EXPIRE_MS;
      setAllLogs((prev) => prev.filter((entry) => entry.receivedAt > expireThreshold));
    };

    cleanupTimerRef.current = setInterval(cleanup, CLEANUP_INTERVAL_MS);

    return () => {
      if (cleanupTimerRef.current !== null) {
        clearInterval(cleanupTimerRef.current);
        cleanupTimerRef.current = null;
      }
    };
  }, []);

  // ---- Wails イベント購読 ---------------------------------------------------
  useWailsEvent<unknown>(WAILS_LOG_EVENT, (payload) => {
    const parsed = logEventSchema.safeParse(payload);
    if (!parsed.success) {
      return;
    }

    const now = Date.now();
    const entry: LogEntryWithMeta = {
      id: `${parsed.data.id ?? now}-${eventCounter++}`,
      level: parsed.data.level ?? 'INFO',
      message: parsed.data.message ?? '',
      timestamp: parsed.data.timestamp ?? new Date().toISOString(),
      attributes: parsed.data.attributes ?? {},
      receivedAt: parseTimestampMs(parsed.data.timestamp),
    };

    setAllLogs((prev) => {
      const next = [entry, ...prev];
      // リングバッファ: 超過分は末尾（古い方）から切り捨て
      return next.length > MAX_LOG_ENTRIES ? next.slice(0, MAX_LOG_ENTRIES) : next;
    });
  });

  // ---- ログクリア（手動） ---------------------------------------------------
  const clearLogs = useCallback(() => {
    setAllLogs([]);
  }, []);

  return {
    /** 現在保持しているすべてのログエントリ（新着順） */
    allLogs: allLogs as LogEntry[],
    /** ログ件数 */
    allLogsCount: allLogs.length,
    /** ログを手動でクリアする */
    clearLogs,
  };
}
