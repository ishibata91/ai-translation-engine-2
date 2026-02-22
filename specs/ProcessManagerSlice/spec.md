# ProcessManager 仕様書

## 概要
システム全体のワークフローを管理し、各垂直スライス間の連携や順序制御をオーケストレーションするコンポーネント。
（※本スライスは現在構想段階であり、段階的に実装される予定です）

## 将来的な責務 (isolate-slice-dtos 関連)

### 1. スライス間連携とDTOマッピングの集中管理
現状、各垂直スライス（TermTranslator, PersonaGen, ContextEngine, SummaryGenerator, Pass2Translator）は、モジュール間の密結合を防ぐために独自の入力用DTOを定義する設計（Anti-Corruption Layer）となっています。
ProcessManagerは、これらのスライスをつなぐハブとして機能します。

- **データ変換**: `loader_slice` 等から取得した共有データ（例: `ExtractedData`）を、各ドメインスライスが個別に要求する専用DTO（例: `TermTranslatorInput`）へ変換（マッピング）する責務を担います。
- **依存の遮断**: 各スライスが `loader_slice` のデータ構造に直接依存しないよう、プロセスマネージャーが中継することで、スライスの完全独立性を担保します。

### 2. その他の制御（予定）
- 各種スライスの呼び出し順序・並列実行の管理
- 全体に対するエラーハンドリングとリトライ制御
- 進捗の統合とUIへの通知ループ
