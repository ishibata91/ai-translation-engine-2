## Why

バックエンドのリファクタリングを安全に進めるための統一ルールと自動品質ゲートが未整備で、実装品質とレビュー観点が人依存になっている。最初に規約とライブラリ導入を確定し、以降の変更を一貫した基準で進められる状態を作る。

## What Changes

- バックエンド標準コーディング規約（命名、依存方向、エラーハンドリング、context伝播、ログ、テスト方針、SRP、公開APIのdoc必須）を定義する。
- このリポジトリのアーキテクチャ原則（Interface-First AIDD / Vertical Slice / DTO分離 / pipeline mapper責務）に沿ったリポジトリ固有規約を定義する。
- de facto standard の品質ライブラリを導入し、lint・整形・脆弱性検査・並行処理リーク検知を自動実行できるようにする。
- 規約違反を検出する実行手順（ローカル実行とCI実行）を定義し、リファクタリング前の必須チェックとして運用する。
- 一般的なチェックスタイル（MUST/SHOULD区分、レビュー時の確認観点、lint違反時の扱い）を規約に含める。
- **BREAKING**: 既存コードが新規約・新lint設定に適合しない場合、修正が必須になる。

## Capabilities

### New Capabilities
- `backend-coding-standards`: バックエンド実装時に必ず従う共通コーディング規約と、リポジトリ固有ルールを定義する。
- `backend-quality-gates`: de facto standard ライブラリに基づく lint / format / vuln check / goroutine leak check を品質ゲートとして定義する。

### Modified Capabilities
- なし

## Impact

- Affected code:
  - `pkg/**`（全バックエンド実装の準拠対象）
  - `openspec/specs/**`（新規 capability spec 追加）
  - CI設定ファイルと開発用スクリプト（新規または更新）
- Affected dependencies:
  - `golangci-lint`（staticcheck, govet, errcheck, revive, gosec など）
  - `goimports`
  - `govulncheck`
  - `go.uber.org/goleak`（テスト補助）
- Affected process:
  - PR前チェック手順とレビュー観点が規約ベースに統一される。
