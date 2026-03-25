---
name: fix-logging
description: AI Translation Engine 2 専用。fix-trace の観測計画に従って一時ログの add/remove を行う。観測用ログ操作に限定して扱う。
---

# Fix Logging

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-logging` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は `fix-trace` が「観測ログが必要」と判断した後の `add` と、最終 accept 後の `remove` cleanup を担当する skill。
責務は観測用ログ操作に限り、サブエージェント起動判断は含めない。

## 使う場面
- `fix-trace` がログ追加を必要と判断し、`fix-direction` 経由で起動された場合
- `fix-direction` が最終 accept 後に一時ログ cleanup を必要とし、`operation: remove` で再起動した場合
- 原因仮説の絞り込みに追加観測が必要なとき

## 入力契約
- `operation: add | remove`
- `fix-trace` が返した原因仮説と観測計画（`operation: add` の起動時パケット）
- 観測を仕込む対象ファイル一覧
- `fix-direction` の state summary
- `log_additions` と対象ファイル一覧（`operation: remove` の cleanup パケット）

## 手順
1. 渡されたパケットから `operation` を確認し、`add` と `remove` 以外ならエラーとして返す。
2. `operation: add` の場合は `fix-trace` の観測計画を確認し、観測ポイントを特定してログを仕込む対象ファイルと行を決定する。
3. `operation: add` の場合は `[fix-trace]` prefix の観測ログを追加し、追加した import / call site を `log_additions` に記録する。
4. `operation: remove` の場合は受け取った `log_additions` を正本として、一時ログと不要になった import を削除し、削除内容を `log_removals` に記録する。
5. 変更したファイル一覧と `log_additions` / `log_removals` を戻り値パケットに含めて返す。
6. `fix-direction` が state summary を更新できるよう、`active_logs` または cleanup 完了状態を戻り値に含める。
7. `fix-direction` へ、`add` では「ログ追加完了・再現待ち」、`remove` では「cleanup 完了」を返す。

## 許可される動作
- 恒久修正は扱わず、観測用ログの add/remove に集中する
- repo 常設 logger を汚さない
- 一時ログの prefix は必ず `[fix-trace]` を使い、add/remove の両 operation で同じ契約を維持する
- フロントエンドのログは必ず `src/lib/logger.ts` の `logger.*` を使う（`console.*` は不可）。
- バックエンドのログは注入済み `*slog.Logger` または `slog.Default()` を使う。
- import と call site を一括削除して戻せる形で追加し、remove は `log_additions` に基づく最小 cleanup だけを行う
- `fix-direction` へ返す内容は `operation`、変更ファイル、`log_additions`、`log_removals`、必要な再現ガイド、state summary 更新材料に限定する

## ログ実装ガイド

### フロントエンド
```ts
import { logger } from '../lib/logger'; // パスは呼び出し元に応じて調整

logger.debug('[fix-trace] <観測ポイント>', { key: 'value' });
logger.info('[fix-trace] <状態>', { phase: 'xxx' });
logger.warn('[fix-trace] <異常値>', { actual: String(val) });
logger.error('[fix-trace] <エラー箇所>', { reason: err.message });
```
出力先: `{リポジトリルート}/logs/YYYY-MM-DD.jsonl`

### バックエンド
```go
// 注入済みロガー（推奨）
logger.DebugContext(ctx, "[fix-trace] <観測ポイント>", slog.String("key", val))

// 注入ロガーがない場合
slog.Default().DebugContext(ctx, "[fix-trace] <観測ポイント>")
```
出力先: stdout（JSON）+ `{リポジトリルート}/logs/YYYY-MM-DD.jsonl`

## 参照資料
- handoff には `references/templates.md` を使う。
