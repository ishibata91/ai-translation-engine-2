## Why

現状のフロントエンド品質ゲートは typecheck、ESLint、Vitest、build までに留まっており、Wails アプリとしての画面起動や主要導線の統合退行を自動検知できない。主要画面の表示崩れや Wails binding 変更に伴う実行時破損を、実装直後に機械検証できる Playwright ベースの E2E 基盤が必要である。

## What Changes

- Playwright を用いたフロントエンド E2E テスト基盤を追加する
- ローカル実行しやすい Playwright 用 script、設定ファイル、初期サンプルテスト、共通 fixture を整備する
- Wails/React 構成に合わせて、最低限のアプリ起動確認と主要画面導線を検証できる実行方式を定義する
- 既存のフロントエンド品質ゲートに Playwright 実行を追加し、作業完了前の必須確認項目として扱えるようにする
- `AGENTS.md` に Playwright をフロントエンド品質ゲートの一部として追記し、AI の標準実行フローへ組み込む

## Capabilities

### New Capabilities
- `playwright-quality-gate`: Playwright による E2E テスト基盤、共通実行設定、初期シナリオ、品質ゲート接続を定義する

### Modified Capabilities
- `frontend-code-quality-guardrails`: フロントエンド品質ゲートに E2E 実行を追加し、`lint:file -> 修正 -> 再実行 -> lint:frontend -> Playwright` の完了条件へ更新する

## Impact

- `frontend/package.json` の scripts と devDependencies
- `frontend/` 配下の Playwright 設定、fixture、E2E テスト配置
- ルートまたは frontend 側の品質ゲート実行導線
- `AGENTS.md` のフロントエンド作業フロー
- 将来的には Wails 実行モードやテスト対象画面の拡張方針にも影響する
