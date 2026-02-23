## 1. リファクタリング準備

- [x] 1.1 `pkg/translator` パッケージ内に `context_engine.go` を作成し、旧 `lore` のロジック（会話ツリー解析等）を移行
- [x] 1.2 `pkg/translator` 内の既存ロジックを整理（`propose.go`, `save.go`, `tag_processor.go` 等へのファイル分割）

## 2. インターフェースと入力DTOの刷新

- [x] 2.1 `translator` スライスの `ProposeJobs` インターフェースを変更し、構築済みの `TranslationRequest` ではなく、生のゲームデータを含むDTOを受け取るように修正
- [x] 2.2 統合後のスライス専用の入力DTO（`TranslatorInput` 等）を定義

## 3. Propose フェーズの実装 (統合フロー)

- [x] 3.1 `ProposeJobs` 内部で `contextEngine` を呼び出し、各レコードに対する文脈構築（前のセリフ特定、話者特定、辞書検索）を実行するように実装
- [x] 3.2 構築された文脈情報をプロンプトテンプレートのプレースホルダーに埋め込む処理を実装
- [x] 3.3 中間データ構造 `TranslationRequest` への外部依存を完全に排除

## 4. オーケストレーター層の修正

- [x] 4.1 `ProcessManager`（または Pipeline）において、`lore` スライスの呼び出し処理を削除
- [x] 4.2 直接 `translator` スライスを呼び出し、ゲームデータを渡すフローに修正

## 5. 動作確認とクリーンアップ

- [x] 5.1 統合後のスライスで、従来通り（あるいはそれ以上）の文脈を保持したプロンプトが生成されることをパラメタライズドテストで確認
- [x] 5.2 不要になった `pkg/lore` パッケージ、および関連する古い DTO 群を削除

## 6. 追加の修正（依存関係と静的解析エラー）

- [x] 6.1 `pkg/pipeline/mapper.go` のインポートサイクル解消とマッピング修正
- [x] 6.2 `pkg/terminology`, `pkg/persona`, `pkg/summary` のインターフェース修正に伴うテストコード等のエラー修正
