## 1. ルール棚卸しと依存準備

- [x] 1.1 `frontend_coding_standards.md` の各規約を `lint で担保する / AI 補完に回す` に分類し、実装対象ルール一覧を確定する
- [x] 1.2 `frontend/package.json` に TSDoc 検査と unused export 検査に必要なデファクト依存を追加する
- [x] 1.3 既存の `eslint.config.js` と `lint:rule3` の役割を整理し、今回の一般化方針に合わせて移行対象を明確にする

## 2. ESLint ルール実装

- [x] 2.1 `frontend/eslint.config.js` に TSDoc 構文検査ルールを追加し、公開 Hook・公開型・公開コンポーネントを対象に必須化する
- [x] 2.2 `frontend/eslint.config.js` に unused export または同等の公開範囲最小化ルールを追加し、既存構成と競合しないように調整する
- [x] 2.3 `pages` から `wailsjs/store` 直接 import 禁止など、Headless Architecture の境界ルールを feature 横断で適用できる形に整理する
- [x] 2.4 既存違反が多いルールは対象ディレクトリや override を使って段階導入できるように設定する

## 3. 部分 lint と AI 修正フロー

- [x] 3.1 `frontend/package.json` に任意ファイルを対象にした `lint:file` 系 script を追加する
- [x] 3.2 `lint:file` が AI に渡しやすい JSON 出力と失敗コードを返すことを確認する
- [x] 3.3 フロント変更時の標準フローを `変更ファイル lint -> AI 修正 -> 再lint -> 最後に lint:frontend` に統一する
- [x] 3.4 lint 化できない規約の AI 補完チェックリストを文書化し、変更ごとに確認できる形にする

## 4. 検証と文書更新

- [x] 4.1 代表的な公開 Hook、公開型、ページコンポーネントで TSDoc ルールが期待どおりに失敗・成功することを確認する
- [x] 4.2 不要 `export`、境界違反 import、変更ファイル単位 lint が期待どおりに検出されることを確認する
- [x] 4.3 `frontend_coding_standards.md` を更新し、lint 化済み規約と AI 補完規約の運用を反映する
- [x] 4.4 作業完了時に `npm run lint:frontend` を実行し、最終品質ゲートとして通過させる
