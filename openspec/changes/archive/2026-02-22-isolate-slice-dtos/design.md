# Design: Isolate Slice DTOs

## Context
現在の実装では、`loader_slice`が定義するDTOを他の複数スライスが直接インポートして利用しています。この状態は、各垂直スライスが外部モジュールに密結合していることを意味し、`loader_slice`の変更が他スライスに直接影響を及ぼすという問題があります。
Vertical Slice Architecture（VSA）の原則に従い、各スライスの完全な独立性を確保するため、依存関係を逆転（あるいは遮断）させる再設計が求められています。

## Goals / Non-Goals
**Goals:**
- `loader_slice`, `term-translator-slice`、`persona-gen-slice`、`context-engine-slice`、`summary-generator-slice`、`pass2-translator-slice` を含む各スライスから、共有DTOへの依存を完全に排除する。
- 共有モジュールとなっている `pkg/domain` フォルダ全体を削除する。
- Consumer-Driven Contracts アプローチ（スライス独自定義）に基づき、各スライスが自身の入力や出力として要求する独自のデータ構造（DTO/インターフェース）を定義する。
- オーケストレーター層（`ProcessManager` など）において、各スライスの独自DTOへのデータマッピング責務を集約する設計とする。

**Non-Goals:**
- `loader_slice` 自体の内部パースロジックやファイル読み込み処理の大幅な変更。
- スライス分割自体の見直しや、UI層の修正。

## Decisions
1. **共有ドメインモデル (`pkg/domain`) の削除**
   - **Rationale**: VSAにおいて、異なるコンテキスト（スライス）間でデータモデルを共有することは推奨されません。共有モデルは変更の理由が複数発生し、各スライスが不要に密結合する原因となります。
   - **Details**: `pkg/domain` ディレクトリを削除し、各スライスがそれぞれ必要な構造体を独自定義するようにします。

2. **スライス固有のDTO定義 (`contract.go`)**
   - **Rationale**: Anti-Corruption Layer（腐敗防止層）の役割として、各スライス（参照側および提供側）は自身が必要とするデータのみを定義します。
   - **Details**: 各スライスの `contract.go` において、入力・出力データとして用いる構造体（例: `TermTranslatorInput`、`LoaderOutput` など）を独自に定義します。DTOの定義場所は `contract.go` とし、スライスの外部インターフェース（窓口）として扱います。

3. **オーケストレーターでのマッピング処理 (今回はスキップ)**
   - **Rationale**: 今回のフェーズでは各スライスの「要求データ構成の宣言（DTOの分離）」のみをスコープとし、オーケストレーターでの実際のマッピング実装は将来のタスクとします。
   - **Details**: まずは各スライス内にDTOを定義し、将来 `ProcessManager` がそれらを呼び出す際のマッピングを担う旨を `/specs/ProcessManagerSlice/` にドキュメントとして残します。

## Risks / Trade-offs
- **[Trade-off] マッピングコードの増加**
  - **Mitigation**: オーケストレーター層に詰め替え処理のボイラープレートコードが増加しますが、これは各スライスの独立性を担保するための必要コストとして許容します。ファクトリ関数やマッパーヘルパーを導入することで可読性を維持します。
- **[Risk] データ構造変更時の追従漏れ**
  - **Mitigation**: `loader_slice`側でデータ構造が変更された際にオーケストレーター層でのコンパイルエラーとして直ちに検知されるため、静的型付け言語（Go）の恩恵により安全に修正が可能です。
