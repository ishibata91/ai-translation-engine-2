## Context

現在、`pkg/loader` は `ExtractedData` をロードするための処理を持っていますが、Interface-First AIDDアーキテクチャ（v2）が要求する「Vertical Slice（縦のコンテキスト）」および「厳格な依存関係の注入（DI）」の要件を完全に満たしていません。
また、ドメインモデルが `pkg/domain/models/models.go` に一極集中しており、ファイルの肥大化が懸念されていました。ドメインモデルそのものを無くすとプロセス全体を管理するProcess Manager（オーケストレーター）からのデータアクセスが困難になるため、モデルの配置場所は維持しつつ、コンテキスト境界に沿ってファイルを分割する必要があります。

## Goals / Non-Goals

**Goals:**
- **Vertical Slice化**: `pkg/loader` 内部のデータロードロジック（直列デコードおよび並行パース）を隠蔽し、外部には `contract.Loader` インターフェースのみを公開・提供する構造にする。
- **DI (Google Wire) 対応**: Loaderモジュールを生成するためのプロバイダ関数 (`ProvideLoader` 等) を定義し、クリーンな依存解決を実現する。
- **ドメインモデルのコンテキスト分割**: `models.go` を維持しつつ、`dialogue.go`, `quest.go`, `entity.go` など関連するデータごとにファイルを分割する。

**Non-Goals:**
- ロード処理そのもののアルゴリズムの抜本的な変更（JSONデコード、パラレル処理の基礎ロジックは既存のものを踏襲する）。
- `pkg/domain/models` 以外の他モジュール（Translate等）の深いリファクタリング。

## Decisions

### 1. ドメインモデル (`pkg/domain/models`) のファイル分割
**決定**: `models.go` を廃止・統合するのではなく、同じ `models` パッケージの中でコンテキストごとに分割する。
**理由**: オーケストレーター層からパース結果である `ExtractedData` 全体を可視化し、次のパイプラインに渡すためには、共通のドメインモデルパッケージが必要です。しかし単一ファイルは保守性が低いため、意味のあるまとまり（対話、クエスト、アイテム・NPC、ベース構造体など）でファイルを分けます。
**分割案**:
- `base.go` (BaseExtractedRecord, ExtractedData コンテナ)
- `dialogue.go` (DialogueResponse, DialogueGroup)
- `quest.go` (Quest, QuestStage, QuestObjective)
- `entity.go` (NPC, Item, Magic)
- `system.go` (Location, Message, SystemRecord, LoadScreen)

### 2. ContractとImplementationの分離
**決定**: `pkg/contract/loader.go` に定義されたインターフェースを引き続き正ととし、`pkg/loader` はあくまでその実装を提供するだけの「詳細」とする。
**理由**: Interface-First方針に従うため。外部からの利用者は `contract.Loader` だけを知っていれば良く、 `loader.go` が内部でどのような並行処理（`ParallelProcessor` 等）をしているかを気にする必要がなくなります。また、テスト時にはMock化が容易になります。

### 3. Google WireによるDIのプロバイダ提供
**決定**: `pkg/loader/provider.go` を新設し、`ProvideLoader() contract.Loader` をエクスポートする。
**理由**: 呼び出し側（MainやProcess Manager）で `wire.NewSet(loader.ProvideLoader)` のように簡単に依存関係を構築できるようにするため。

## Risks / Trade-offs

- **[Risk]** ドメインモデルのファイル分割により、一括で参照していたコードの探索がわずかに手間になる可能性がある。
  - **Mitigation**: Goのパッケージとしては同一（`package models`）であるため、利用側のインポート文（`import "github.com/.../models"`）は変更されず、IDE上で問題なく定義ジャンプが可能です。呼び出し側への影響はありません。
- **[Risk]** Interfaceを通した処理により間接参照が増える。
  - **Mitigation**: v2のアーキテクチャ原則における保守性・テスト容易性のメリットが上回るため許容します。性能上のオーバーヘッドは微小です。
