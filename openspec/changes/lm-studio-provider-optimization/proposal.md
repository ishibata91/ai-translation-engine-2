## Why

現行のローカルLLMプロバイダ実装はプロバイダ固有機能（モデル一覧取得、ロード/アンロード、構造化出力）を前提にした設計になっておらず、モデル更新追従と実行再開性を同時に満たせていない。LM Studio の公式インターフェースに合わせて llm モジュールを再定義し、queue 連携で中断・再開時も一貫した実行保証を提供する必要がある。

## What Changes

- ローカルLLMプロバイダの名称・責務を `LM Studio` プロバイダへ統一し、設定・実装・ドキュメントの呼称を置換する。
- `llm` モジュール全体に、全プロバイダ共通で利用できる「モデルリスト取得インターフェース」を追加する。
- `LM Studio` プロバイダで、リクエスト開始時のモデルロード（`/api/v0/load`）を必須化する。
- `LM Studio` プロバイダで、中断時および処理完了時のモデルアンロード（`/api/v0/unload`）を必須化する。
- `llm` パッケージ契約として、Structured Output は全プロバイダでサポートすべき要件を明記する（今回の実装対象は `LM Studio` のみ）。
- `LM Studio` プロバイダで、OpenAI互換 Structured Output を先行実装する。
- `specs/queue` と連携し、LLM実行の中断・再開時にモデル状態とジョブ状態の整合を保つ再開性要件を追加する。
- **BREAKING**: 既存の `local`/`local-llm` 相当のプロバイダ識別子・設定キーは `lmstudio` 系へ移行する。

## Capabilities

### New Capabilities
- `llm-model-discovery`: llm モジュール共通のモデル一覧取得インターフェースと、プロバイダ別モデル列挙の標準契約を定義する。
- `lmstudio-provider-lifecycle`: LM Studio 向けのモデルロード/アンロード、Structured Output先行実装、実行ライフサイクル要件を定義する。

### Modified Capabilities
- `queue`: LLMジョブの中断・再開時に、モデルロード状態と処理状態を復元可能にする再開性要件へ拡張する。

## Impact

- Affected specs: `specs/llm/*`, `specs/queue/spec.md`
- Affected code: `pkg/llm/*`（プロバイダIF、実装、エラー処理、設定解決）、queue 連携箇所
- Affected APIs: LM Studio REST (`/api/v0/models`, `/api/v0/load`, `/api/v0/unload`) と OpenAI互換 Structured Output
- Contract note: Structured Output は `pkg/llm` のプロバイダ共通契約として定義し、今回チェンジで実装保証するのは `LM Studio` のみ
- Dependencies/standards:
  - 既存 `go-openai`（デファクト）を Structured Output 呼び出し経路で継続活用
  - LM Studio Developer Docs 準拠のHTTP契約へ統一
- Migration: 設定キー/プロバイダ名の互換レイヤーまたは移行手順の提供が必要


