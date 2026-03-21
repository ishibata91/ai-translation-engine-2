/**
 * logger — フロントエンド発ログをバックエンド経由で logs/ に書き込むユーティリティ
 *
 * 使い方:
 *   import { logger } from '../lib/logger';
 *
 *   logger.info('設定を読み込みました', { namespace: 'model-settings' });
 *   logger.warn('モデルが未選択です');
 *   logger.error('API 呼び出し失敗', { endpoint: url, reason: err.message });
 *   logger.debug('レンダリング', { component: 'Dashboard' });
 *
 * 注意:
 *   - WriteLog は Wails バインディング（非同期 IPC）のため fire-and-forget で投げる。
 *   - Wails 未起動時（ブラウザ単体テスト等）はコンソールにフォールバックする。
 */

import {WriteLog} from '../wailsjs/go/controller/TelemetryController';

// ---- 型定義 -----------------------------------------------------------------

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

export type LogAttrs = Record<string, string>;

// ---- 内部ヘルパー -----------------------------------------------------------

/**
 * Wails バインディングが利用可能かどうかを確認する。
 * ブラウザ単体（非 Wails）環境では window.go が未定義になる。
 */
function isWailsAvailable(): boolean {
    return typeof window !== 'undefined' && 'go' in window;
}

/**
 * ログをバックエンドに送信する。
 * Wails が未初期化の場合はコンソールにフォールバックする。
 */
function send(level: LogLevel, message: string, attrs?: LogAttrs): void {
    const timestamp = new Date().toISOString();

    if (!isWailsAvailable()) {
        // 開発・テスト環境用フォールバック（コンソールのみ）
        const out = { level, message, timestamp, attrs };
        switch (level) {
            case 'DEBUG': console.debug('[frontend]', out); break;
            case 'INFO':  console.info('[frontend]', out);  break;
            case 'WARN':  console.warn('[frontend]', out);  break;
            case 'ERROR': console.error('[frontend]', out); break;
        }
        return;
    }

    // fire-and-forget: バックエンド書き込み失敗はコンソールに落とすのみ
    WriteLog({ level, message, timestamp, attrs: attrs ?? {} }).catch((err: unknown) => {
        console.error('[logger] WriteLog failed:', err);
    });
}

// ---- 公開 API ---------------------------------------------------------------

export const logger = {
    /**
     * DEBUG レベルのログを出力する。
     * 開発中の詳細トレースに使用する。
     */
    debug(message: string, attrs?: LogAttrs): void {
        send('DEBUG', message, attrs);
    },

    /**
     * INFO レベルのログを出力する。
     * 通常の操作ログに使用する。
     */
    info(message: string, attrs?: LogAttrs): void {
        send('INFO', message, attrs);
    },

    /**
     * WARN レベルのログを出力する。
     * 注意が必要だが処理継続可能な事象に使用する。
     */
    warn(message: string, attrs?: LogAttrs): void {
        send('WARN', message, attrs);
    },

    /**
     * ERROR レベルのログを出力する。
     * 操作失敗・例外など重篤な事象に使用する。
     */
    error(message: string, attrs?: LogAttrs): void {
        send('ERROR', message, attrs);
    },
} as const;
