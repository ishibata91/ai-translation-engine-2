## MODIFIED Requirements

### Requirement: 用語翻訳対象データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライス(LoaderSlice等)のDTOへの依存を排除するため。
**Migration**: 外部の `ExtractedData` などを直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（例: `TermTranslatorInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化と翻訳処理の開始
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`TermTranslatorInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いて用語翻訳リクエストを生成できること
