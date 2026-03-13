## 1. Playwright 基盤セットアップ

- [x] 1.1 `frontend/package.json` に Playwright の devDependencies と実行 script を追加する
- [x] 1.2 `frontend` 配下に Playwright 設定ファイルを追加し、Vite アプリを起動して E2E を実行できるようにする
- [x] 1.3 Playwright のテスト配置ディレクトリと共通初期化ファイルを作成する

## 2. 共通 fixture と最小 E2E 導線

- [x] 2.1 ベース URL への遷移と主要レイアウト待機を共通化する fixture または helper を追加する
- [x] 2.2 HashRouter 前提で現在ルートや主要画面の識別を安定確認できる helper を整備する
- [x] 2.3 Wails 依存が強い処理を避けつつ、主要画面導線だけを検証対象にする前提をテストコードへ反映する

## 3. 最小品質ゲートシナリオの実装

- [x] 3.1 アプリ起動時にヘッダー、サイドバー、メインコンテンツ領域が表示されることを確認する E2E を追加する
- [x] 3.2 ダッシュボード初期表示を確認する E2E を追加する
- [x] 3.3 `DictionaryBuilder` への遷移と画面識別を確認する E2E を追加する
- [x] 3.4 `MasterPersona` への遷移と画面識別を確認する E2E を追加する

## 4. 品質ゲート導線と文書更新

- [x] 4.1 `frontend` の品質ゲート script を更新し、`lint:frontend` 後に Playwright を実行できるようにする
- [x] 4.2 `AGENTS.md` にフロント変更時の標準フローとして `lint:file -> 修正 -> 再実行 -> lint:frontend -> Playwright` を追記する
- [x] 4.3 ローカルで `lint:file`、`lint:frontend`、Playwright を順に実行し、品質ゲートとして成立することを確認する
