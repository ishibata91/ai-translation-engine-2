# データスキーマ & ローダー テスト仕様書

> Interface-First テスト設計
> Phase 1: Data Foundation
> ExtractedData および Loader の正常性検証

---

## 1. ドメインモデル & ヘルパーメソッド テスト

純粋なデータ構造としての整合性と、付随するヘルパーメソッドの動作を検証する。

### 1.1 NPC

| ID     | テスト名             | 入力          | 期待  |
| ------ | -------------------- | ------------- | ----- |
| NPC-01 | IsFemale (Female)    | Sex="Female"  | true  |
| NPC-02 | IsFemale (LowerCase) | Sex="female"  | true  |
| NPC-03 | IsFemale (Male)      | Sex="Male"    | false |
| NPC-04 | IsFemale (Unknown)   | Sex="Unknown" | false |

### 1.2 Item (Type Hinting)

※ 将来的な拡張（TypeHintメソッド等）を見越したカテゴリ。現在は特になし。

---

## 2. ローダー機能テスト (LoadExtractedJSON)

入力JSONファイルから ExtractedData 構造体へのマッピングを検証する。

### 2.1 正常系ロード

| ID     | テスト名           | 入力               | 期待                                 |
| ------ | ------------------ | ------------------ | ------------------------------------ |
| LDR-01 | 基本ロード (UTF-8) | 有効なJSON (UTF-8) | エラーなし, 全フィールドにデータ格納 |
| LDR-02 | BOM付きUTF-8       | UTF-8-SIG          | エラーなし, 正常パース               |
| LDR-03 | Shift-JIS自動検知  | SJISエンコードJSON | エラーなし, 文字化けなし             |
| LDR-04 | 部分的データ       | "quests"のみのJSON | questsのみ格納, 他は空スライス       |

### 2.2 エラーハンドリング

| ID     | テスト名     | 入力              | 期待                        |
| ------ | ------------ | ----------------- | --------------------------- |
| ERR-01 | ファイル不在 | 存在しないパス    | FileNotFoundError           |
| ERR-02 | 不正JSON     | 壊れたJSON構文    | JSONDecodeError             |
| ERR-03 | 空ファイル   | サイズ0のファイル | Evaluated as Empty or Error |

### 2.3 データ正規化 & フィルタリング

| ID    | テスト名     | 入力              | 期待                               |
| ----- | ------------ | ----------------- | ---------------------------------- |
| NR-01 | EditorID抽出 | "Name [VTYP:123]" | EditorID="Name"                    |
| NR-02 | 日本語判定   | "こんにちは"      | IsJapanese=true                    |
| NR-03 | 日本語除外   | 日本語名NPC       | ExtractedDataに含まれない (Filter) |
| NR-04 | 英語保持     | 英語名NPC         | ExtractedDataに含まれる            |

---

## 3. 統合テスト (End-to-End)

| ID     | テスト名       | 内容                                                           |
| ------ | -------------- | -------------------------------------------------------------- |
| E2E-01 | 実データロード | 実JSON → LoadExtractedJSON → メモリ展開確認                    |
| E2E-02 | 大規模データ   | 数万件レコードのロード時間計測 (Go並列化ベンチマーク)          |
| E2E-03 | 破損耐性       | 不正レコード混入JSON → Partial Loading (Validなものだけロード) |
