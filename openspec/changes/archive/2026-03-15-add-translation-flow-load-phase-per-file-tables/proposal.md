## Why

現在の `TranslationFlow` は用語フェーズから始まるため、抽出データを読み込んで内容を確認する入口がなく、複数ファイルを扱う翻訳プロジェクトの開始手順が UI 上で完結していない。さらに、ロード済みデータを UI の一時 state だけで持つと後続フェーズへの受け渡しや再開時の復元ができないため、要件定義と `artifact` 境界に沿ってパース済みデータを保存する必要がある。

## What Changes

- `TranslationFlow` の先頭にデータロードフェーズを追加し、既存の工程タブとステップ表示に組み込む。
- データロードフェーズで複数ファイルを選択できる入力 UI を追加する。
- 選択済みファイルをパースし、その結果を後続フェーズが参照できるよう `artifact` DB に保存する。
- `artifact` に保存した翻訳対象データを `DataTable` で一覧表示できるようにする。
- 一覧表示は全ファイルを 1 つの表へ結合せず、ファイルごとに独立したテーブルとして表示する。
- 各ファイルのテーブルは折りたたみ可能にし、必要なファイルだけ展開して確認できるようにする。

## Capabilities

### New Capabilities
- `frontend/translation-flow-data-load-ui`: 翻訳フローの先頭で複数ファイルを読み込み、artifact に保存されたパース済みデータをファイル単位の折りたたみテーブルで確認できる UI フローを定義する
- `workflow/translation-flow-data-load`: 翻訳フローのデータロード phase 順序と artifact handoff を定義する

### Modified Capabilities
- `artifact/shared-handoff`: 翻訳フローのロード結果を後続フェーズへ受け渡せるよう、`task_id` 基点の構造化テーブルへパース済みデータを保存・検索する要件を追加する

## Impact

- フロントエンド: `frontend/src/pages/TranslationFlow.tsx`、新規または既存の feature hook、`frontend/src/components/DataTable.tsx`、必要に応じて翻訳フロー配下コンポーネント
- バックエンド: translation flow のロード要求を受ける controller / workflow、parser 出力を `artifact.db` の専用テーブルへ保存する repository、preview 読み出し API
- OpenSpec: `openspec/changes/add-translation-flow-load-phase-per-file-tables/specs/frontend/translation-flow-data-load-ui/spec.md`、`openspec/changes/add-translation-flow-load-phase-per-file-tables/specs/workflow/translation-flow-data-load/spec.md`、`openspec/changes/add-translation-flow-load-phase-per-file-tables/specs/artifact/shared-handoff/spec.md`
- API / データ境界: パース済みデータは `artifact.db` の `task_id` 基点の構造化テーブル群へ保存し、translation flow はその保存結果を参照する
- ER 図影響: 新規 DB ファイル追加は行わないが、`database_erd.md` の artifact セクションに translation flow 用の task / file / section テーブル群を追加する
- 依存: 既存の React / TanStack Table、task ストア、parser スライスを継続利用し、新規ライブラリ追加は行わない
