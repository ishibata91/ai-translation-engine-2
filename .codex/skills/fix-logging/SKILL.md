---
name: fix-logging
description: AI Translation Engine 2 専用。fix-trace から「ログを仕込む必要がある」と判断が返ってきたとき、観測用の一時ログを実装する。恒久修正は行わない。
---

# Fix Logging

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-logging` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は `fix-trace` が「観測ログが必要」と判断した後に起動され、観測用の一時ログをコードへ追加する skill。
恒久修正は行わない。他スキルのサブエージェントを呼び出さない。

## 使う場面
- `fix-trace` がログ追加を必要と判断し、`fix-direction` 経由で起動された場合
- 原因仮説の絞り込みに追加観測が必要なとき

## 入力契約
- `fix-trace` が返した原因仮説と観測計画（起動時のパケットとして受け取る）
- 観測を仕込む対象ファイル一覧

## 手順
1. 渡されたパケットから `fix-trace` の観測計画を確認する。
2. 観測ポイントを特定し、ログを仕込む対象ファイルと行を決定する。
3. 観測ログを追加する（import / call site の追加）。
4. 追加したログの一覧を戻り値のパケットに含めて返す。
5. `fix-direction` へ「ログ追加完了・再現待ち」を返す。

## 原則
- 恒久修正は行わない
- repo 常設 logger を汚さない
- 一時ログは後で一括削除しやすいよう、ログメッセージにtemp接頭辞をつける。例: `[temp_fix-trace] <message>`
- フロントエンドのログは必ず `src/lib/logger.ts` の `logger.*` を使う（`console.*` は不可）。
- バックエンドのログは注入済み `*slog.Logger` または `slog.Default()` を使う。
- import と call site を一括削除して戻せる形で追加する
- `fix-direction` へ返す内容は「追加したログ一覧」と「再現に必要な操作ガイド」だけとする

## ログ実装ガイド

### フロントエンド
```ts
import { logger } from '../lib/logger'; // パスは呼び出し元に応じて調整

logger.debug('[temp_fix-trace] <観測ポイント>', { key: 'value' });
logger.info('[temp_fix-trace] <状態>', { phase: 'xxx' });
logger.warn('[temp_fix-trace] <異常値>', { actual: String(val) });
logger.error('[temp_fix-trace] <エラー箇所>', { reason: err.message });
```
出力先: `{リポジトリルート}/logs/YYYY-MM-DD.jsonl`

### バックエンド
```go
// 注入済みロガー（推奨）
logger.DebugContext(ctx, "[temp_fix-trace] <観測ポイント>", slog.String("key", val))

// 注入ロガーがない場合
slog.Default().DebugContext(ctx, "[temp_fix-trace] <観測ポイント>")
```
出力先: stdout（JSON）+ `{リポジトリルート}/logs/YYYY-MM-DD.jsonl`

## 参照資料
- handoff には `references/templates.md` を使う。
