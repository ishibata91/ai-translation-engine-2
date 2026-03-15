## 1. Artifact ストレージ整備

- [x] 1.1 `artifact.db` に `translation_input_tasks`、`translation_input_files`、parser section 対応テーブル群を追加する migration / schema 初期化を実装する
- [x] 1.2 `task_id` を親キーとしてファイル親行と `preview_row_count` を保存できる repository を実装する
- [x] 1.3 `DialogueGroup -> DialogueResponse`、`Quest -> QuestStage / QuestObjective` の親子保存ロジックを実装する
- [x] 1.4 preview 用に全 section を 50 行単位で横断取得できる query を repository に実装する
- [x] 1.5 変更したバックエンドファイルに対して `npm run backend:lint:file -- <file...>` を実行し、指摘を潰して再実行する

## 2. Workflow / Controller 連携

- [x] 2.1 翻訳プロジェクト task の `task_id` を受け取り、複数ファイルを parser へ流して artifact へ保存する workflow を実装する
- [x] 2.2 ロード済みファイル一覧と file ごとの preview 行を返す backend API / controller を実装する
- [x] 2.3 backend API 追加に伴う Wails binding と adapter 境界を更新する
- [x] 2.4 backend のテーブル駆動テストを追加し、保存・再読込・50 行ページング・全 section preview を検証する
- [x] 2.5 `npm run lint:backend` と必要に応じて `go test ./pkg/...` を実行して backend 品質ゲートを通す

## 3. TranslationFlow フロント実装

- [x] 3.1 `useTranslationFlow` と関連型を追加し、`TranslationFlow.tsx` のタブ状態・ロード状態・ページング状態を hook へ移す
- [x] 3.2 `LoadPanel.tsx` を追加し、複数ファイル選択 UI、選択済みファイル一覧、ロード実行操作を実装する
- [x] 3.3 backend から返る file / preview データを `LoadedTranslationFile` と `TranslationTargetRow` へ整形する adapter を実装する
- [x] 3.4 `DataTable` を使ってファイルごとの折りたたみテーブルと 50 行ページング UI を実装する
- [x] 3.5 `TranslationFlow` のステップ表示とタブ表示に `データロード` フェーズを追加し、ロード完了後に `用語` へ遷移できるようにする
- [x] 3.6 変更したフロントファイルに対して `npm run lint:file -- <file...>` を実行し、修正後に再実行する

## 4. 統合確認

- [x] 4.1 main spec と ER 図の記述が実装内容と乖離していないか確認し、必要なら追従修正する
- [x] 4.2 `npm run typecheck` を実行してフロントの型整合性を確認する
- [x] 4.3 `npm run lint:frontend` を実行してフロント品質ゲートを通す
- [x] 4.4 Playwright E2E で複数ファイル選択、ファイル別テーブル表示、折りたたみ、ページ切り替えを確認する
- [x] 4.5 `review.md` の品質ゲートと完了条件を更新する
