# Tasks: Frontend Coding Standards & Refactor Preparation

## 1. 規約定義の確定

- [x] 1.1 `openspec/specs/frontend_coding_standards.md` の内容レビュー
- [x] 1.2 規約の適用対象（`frontend/src` 全体）を合意
- [x] 1.3 例外ルール（暫定許容箇所）の初期リストを作成

## 2. Lint/Format 基盤整備

- [x] 2.1 `eslint` と `typescript-eslint` を導入
- [x] 2.2 `eslint-plugin-import` を導入し `import/no-restricted-paths` を設定
- [x] 2.3 `pages` -> `wailsjs` / `store` の直接 import 禁止ルールを追加
- [x] 2.4 `prettier` と `eslint-config-prettier` を導入
- [x] 2.5 `npm run lint` / `npm run typecheck` スクリプトを整備

## 3. 既存コードの規約適用（第一弾）

- [x] 3.1 `DictionaryBuilder` の UI責務とロジック責務を再点検し、ページを薄く保つ
- [x] 3.2 `useDictionaryBuilder` の戻り値を state/action 単位で整理
- [x] 3.3 Wailsレスポンス整形処理を adapter 関数へ抽出

## 4. 既存コードの規約適用（第二弾）

- [x] 4.1 `useMasterPersona` を責務別に分割（設定永続化/タスク監視/NPC取得）
- [x] 4.2 `Events.EventsOn` の購読処理を共通Hook化
- [x] 4.3 `as any` 利用箇所を `unknown + 型ガード` へ置換

## 5. テスト・検証基盤

- [x] 5.1 `vitest` + Testing Library を導入
- [x] 5.2 主要 feature hook の振る舞いテストを追加
- [x] 5.3 `npm run test` / `npm run build` を品質ゲートに追加

## 6. 完了条件

- [x] 6.1 CI またはローカルで `typecheck/lint/test/build` が全通過
- [x] 6.2 規約違反 import が 0 件
- [x] 6.3 規約MDと実装状態の差分を変更ログに記録



