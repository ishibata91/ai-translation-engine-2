## MODIFIED Requirements

### Requirement: ペルソナ生成データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライス(LoaderSlice等)のDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（例: `PersonaGenInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化と生成処理の開始
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`PersonaGenInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いてペルソナ生成処理を完結できること
