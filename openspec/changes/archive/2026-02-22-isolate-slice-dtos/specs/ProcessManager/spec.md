## MODIFIED Requirements

### Requirement: スライス間連携とDTOマッピングの集中管理（設計のみ）
**Reason**: 各垂直スライス（TermTranslator, PersonaGen, ContextEngine, SummaryGenerator, Pass2Translator）が `loader_slice` のDTOに依存しないようにするため、連携のハブとなるオーケストレーター層でマッピングの責務を引き受ける必要があるため。
**Migration**: 今回の実装では、各ドメインスライスが専用の入力DTOを宣言する（要求データ構成宣言）のみとし、実際のマッピング処理実装は見送ります。代わりに、`specs/ProcessManagerSlice/spec.md` を新規作成し、将来的に上記のマッピング処理を行う責務について明記します。

#### Scenario: 将来的な連携責務のドキュメント化
- **WHEN** プロジェクト仕様として、オーケストレーターの役割を明確にする必要がある場合
- **THEN** `ProcessManager` の仕様書にて、各スライス専用DTOへの変換と連携の責務が定められていること

#### Specifications: ProcessManagerのDTOマッピングの責務
- `ProcessManager` (またはそれに類するオーケストレーター層) は、各スライス間のデータ受け渡し時にDTOのマッピング（詰め替え）を行う責任を持ちます。
- 各垂直スライス（`loader_slice`, `term-translator-slice`, `persona-gen-slice`, `context-engine-slice`, `summary-generator-slice`, `pass2-translator-slice`）は、それぞれ自らが要求する入力データ構造（`*Input`）と出力データ構造（`*Result`, `*Output` など）を独自に定義しています。
- `ProcessManager`は、例えば `loader_slice` が出力した `LoaderOutput` を受け取り、それを `TermTranslatorInput` や `PersonaGenInput` など、後続のスライスが要求する型に変換（マッピング）してから渡します。
- これにより、各スライスは他のスライスのデータ構造に依存せず、真の垂直分割（Vertical Slice Architecture）および腐敗防止層（Anti-Corruption Layer）の概念を実現します。
