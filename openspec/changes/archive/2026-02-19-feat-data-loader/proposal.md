# Data Loader (Phase 1) Proposal

## Why

現在のPython製データローダーはシングルスレッド処理であり、大量データのロード時にパフォーマンスがボトルネックとなっている。また、動的型付けによる保守性の問題や、進捗状況の可視化が困難であるという課題がある。
これらを解消するため、Go言語による並列処理と静的型付け導入、および「Interface-First AIDD」アーキテクチャへの準拠が必要である。

## What Changes

*   **言語移行**: PythonからGoへ移行し、データ処理を高速化・堅牢化する。
*   **アーキテクチャ刷新**: `Interface-First AIDD` に従い、Contract（構造体定義）とImplementation（ローダーロジック）を分離する。
*   **並列処理**: Goroutineを活用し、JSONパースと正規化処理を並列化する。
*   **文字コード対応**: UTF-8, BOM付き, CP1252, Shift-JIS を自動判別してロードする機能を実装する。

## Capabilities

### New Capabilities
- `LoadExtractedJSON`: 抽出されたJSONファイルを高速に読み込み、メモリ上の構造体(`ExtractedData`)に展開する。
- `EncodingDetection`: ファイルの文字コードを自動判別し、適切なデコーダで読み込む。
- `ParallelParsing`: 会話データ、クエストデータなどを並列にパースし、ロード時間を短縮する。

## Impact

- `pkg/domain/models/`: 新規作成。ドメインモデル（構造体）の定義。
- `pkg/infrastructure/loader/`: 新規作成。データロードロジックの実装。
- `cmd/loader/`: (Optional) 動作確認用のCLIエントリーポイント。

---

このプロポーザルは `openspec/specs/refactoring_strategy.md` および `openspec/specs/data_loader/` 以下の設計資料に基づいています。
