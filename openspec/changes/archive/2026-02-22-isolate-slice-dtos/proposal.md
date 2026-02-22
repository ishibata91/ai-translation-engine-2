# Proposal: Isolate Slice DTOs

## Motivation
現状、各垂直スライス同士がDTOを介して依存し合ってしまっており、スライスの完全な独立性が損なわれている状態です。各スライスが他のスライスや外部のデータ構造に依存すると、変更の余波が広がりやすく、保守性が低下します。
この問題を解決し、アーキテクチャの原則である「スライスの独立性（Anti-Corruption Layer）」を担保するために、データ連携の仕組みをリファクタリングする必要があります。

## Impact
Consumer-Driven Contracts（スライス独自定義）アプローチを採用し、各スライス（参照側）および `loader_slice` が「自身が必要とするデータ構造（あるいは出力構造）」を自身のパッケージ内に独自定義する構成に変更します。
これにより、共有モジュールとなっていた `pkg/domain` フォルダ全体を削除し、システムからグローバルなドメインモデルを排除します。
上位のオーケストレーター（ProcessManagerなど）が `loader_slice` の結果を受け取り、それを各スライス専用のDTOにマッピングして渡すようにします。
既存のソースコードの連携部分が変更されますが、スライス単体としては外部依存が排除され、完全な VSA (Vertical Slice Architecture) へ移行します。

## Capabilities

### New Capabilities
- 該当なし（既存処理のリファクタリングが主目的のため、新規のドメイン機能としてのCapability追加はなし）

### Modified Capabilities
- `loader_slice`: 共有の `pkg/domain` への依存を排除し、独自の出力DTOを定義するように変更。
- `term-translator-slice`: 外部DTOへの依存を排除し、パッケージ独自のDTOで用語連携を行うように変更。
- `persona-gen-slice`: 外部DTOへの依存を排除し、独自のDTOを受け取るように変更。
- `context-engine-slice`: 独自DTOによる連携へ変更。
- `summary-generator-slice`: 独自DTOによる連携へ変更。
- `pass2-translator-slice`: 外部DTOへの依存を排除し、独自のDTOを受け取るように変更。
- `process-manager` (オーケストレーターレイヤー): 今回の実装では対象外。将来的に各スライスのDTOへマッピングを行う責務について `/specs/ProcessManagerSlice/spec.md` に明記するのみに留める。

### Removed Capabilities
- **グローバルドメインモデル**: `pkg/domain` 全体を削除。共有データ構造は VSA の原則に反するため廃止。
