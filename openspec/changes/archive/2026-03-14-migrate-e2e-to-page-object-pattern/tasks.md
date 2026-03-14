## 1. PageObject 基盤の導入

- [x] 1.1 `frontend/src/e2e/page-objects` 配下に BasePO と共通部品PO（Layout/Sidebar/Header）の雛形を作成する
- [x] 1.2 画面PO（Dashboard / DictionaryBuilder / MasterPersona）を作成し、画面遷移と可視確認 API を定義する
- [x] 1.3 runtime 例外検知ロジックを BasePO へ移し、各 PO 操作後に共通検証できるようにする

## 2. Fixture の再設計と注入方式への移行

- [x] 2.1 `app.fixture.ts` の単一 `AppHarness` を廃止し、Wails mock 初期化 + PageObject 注入構成へ変更する
- [x] 2.2 fixture から操作/検証ロジックを削除し、責務を初期化と依存提供に限定する
- [x] 2.3 既存 helper（`routes.ts` / `wails-mock.ts`）の再利用箇所を整理し、PageObject から一貫して利用できるようにする

## 3. Spec の PageObject 化

- [x] 3.1 `app-shell.spec.ts` を PageObject API 呼び出し中心へ書き換える
- [x] 3.2 spec から直接 locator 参照（`page.getBy*` / `page.locator`）を排除する
- [x] 3.3 既存4シナリオ（起動表示、ダッシュボード、辞書遷移、マスターペルソナ遷移）の検証観点が維持されることを確認する

## 4. 旧構成の撤去と品質ゲート確認

- [x] 4.1 旧 harness 由来の不要コード・重複ロジックを削除し、`src/e2e` 配下を新構成に統一する
- [x] 4.2 変更ファイルに対して `npm run lint:file -- <file...>` を実行し、違反を解消する
- [x] 4.3 `npm run lint:frontend` を実行し、フロント品質ゲートを通過させる
- [x] 4.4 `npm run e2e` を実行し、4シナリオがすべて通過することを確認する
