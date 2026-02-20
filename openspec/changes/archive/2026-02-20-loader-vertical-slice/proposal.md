## Why

現在のAI翻訳エンジンv2では、PythonからGoへの移行（Interface-First AIDDアーキテクチャ準拠）を進めています。`openspec/specs/refactoring_strategy.md` のステップ5に従い、エラーハンドリングやDI（Dependency Injection）が不十分だった初期実装の `pkg/loader` を、完全に自律した「Vertical Slice（縦のコンテキスト）」として再構築します。これにより、インターフェース（Contract）と実装（Implementation）を厳密に分離し、保守性と並行処理の安全性を向上させます。

## What Changes

- `pkg/loader` ディレクトリ内の既存コードをVertical Sliceアーキテクチャに適合するようにリファクタリングします。
- **モデルの維持とコンテキスト分割**: `pkg/domain/models` はプロセスマネージャーから抽出データを透過的に扱うために廃止せず維持します。ただし、現在巨大な単一ファイル（`models.go`）となっているため、これらをコンテキスト（Dialogue, Quest, Entity, System等）ごとに小ファイルに分割し、責務を明確にします。
- **Contractの定義**: `pkg/contract` パッケージを通して、モジュール間の契約（Loaderインターフェース等）を明確に定義し、具象実装への直接依存を排除します。
- **DIの導入**: `google/wire` を使用した依存性の注入を前提としたプロバイダ関数（New関数）を整備し、自律的に初期化されるモジュールにします。
- **実装の隠蔽**: 具体的なデータ変換ロジック（Phase 1: Serial Decode, Phase 2: Parallel Process等）は構造体の背後に隠蔽し、Process Manager（UI/オーケストレーター）からの呼び出しを極力シンプルにします。

## Capabilities

### New Capabilities
- `loader-slice-architecture`: Loaderモジュールを自律したVertical Sliceとして連携・初期化するDI/Wireアーキテクチャ。
- `domain-models-split`: `pkg/domain/models` 内の構造体定義をコンテキストベースでファイル分割する管理手法。

### Modified Capabilities
- `data_loader`: 既存のLoader処理のインターフェース契約やDIのプロバイダとしての枠組みに関する変更（処理自体の振る舞いは同一だが、構造化の見直し）。

## Impact

- `pkg/domain/models/` 配下のファイル構造（`models.go` が複数ファイルに分割される）
- `pkg/loader/loader.go` および関連する内部ファイルの実装の隠蔽化
- `pkg/contract/` などのインターフェース定義の網羅性の向上
- Process Managerに相当する起動側・テストコードのDI（Wire）依存の変更
