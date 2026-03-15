# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここでは `AGENTS.md` に定義された品質ゲート確認と、ユーザが提示した完了条件だけを扱う。
共通観点は `openspec/review_standard.md` を前提とする。

## 1. ユーザが出した完了条件

- 完了条件 1: ダッシュボードにタスク削除ボタンを追加する
- 完了条件 2: タスク削除時に task ごとの中間成果物も削除する
- 完了条件 3: OpenSpec の設計 artifacts を design まで作成する

## 2. 品質ゲート確認

該当するものだけ記入する。

### Backend

- [ ] 変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行した
- [ ] `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進めた
- [ ] 作業中または完了前に `npm run lint:backend` を実行した
- [ ] 必要に応じて `npm run backend:check` または `npm run backend:watch` で品質確認した

### Frontend

- [ ] 変更中ファイルに対して `npm run lint:file -- <file...>` を逐次実行した
- [ ] `lint:file -> 修正 -> 再実行 -> 最後に lint:frontend` の順で進めた
- [ ] 作業完了前に `npm run lint:frontend` を実行した

## 3. 実行メモ

- 実行したコマンド:
- 未実行の品質ゲートと理由:
- レビュー時の補足:
