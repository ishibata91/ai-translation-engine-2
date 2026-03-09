## Why

公開メソッドの責務過多は `backend_coding_standards.md` の重要なレビュー観点だが、行数やネスト深度だけでは設計品質を十分に判定できない。高コストな設計 lint を安易に blocking 導入する前に、実装可否・誤検知率・MVP の成立条件を整理する必要がある。

## What Changes

- 公開メソッドの責務過多を検出する設計 lint が品質ゲートとして成立するかを調査する。
- 状態解決、永続化、ログ、goroutine 起動など複数責務を 1 メソッドで抱える代表ケースを分析対象として整理する。
- ルール化する場合の最小 MVP、blocking 導入条件、誤検知が高い場合にレビュー運用へ残す判断基準を定義する。
- 調査結果を `tools/backendquality` へ統合可能な設計方針としてまとめる。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `backend-quality-gates`: 高コストな設計 lint を blocking 導入する前に、成立性評価と導入条件を定義する。

## Impact

- Affected code:
  - `tools/backendquality/**`
  - `pkg/**` の代表的な公開メソッド
- Affected dependencies:
  - 調査段階では `golang.org/x/tools/go/analysis` と既存標準ライブラリを前提に比較する
- Affected process:
  - 設計 lint を導入するか、レビュー運用へ残すかの判断基準が明文化される
