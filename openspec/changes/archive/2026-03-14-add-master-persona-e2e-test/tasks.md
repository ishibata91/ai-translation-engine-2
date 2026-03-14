## 1. OpenSpec とテストデータの準備

- [x] 1.1 `openspec/specs/e2e-required-scenarios/master-persona/spec.md` を main specs 側へ反映する方針を確認し、変更対象の必須シナリオと実装対象ファイルを対応付ける
- [x] 1.2 `frontend/src/e2e/fixtures` 配下に `MasterPersona` 用 mock データを追加し、NPC 一覧、プロンプト初期値、モデル設定初期値、JSON パス、タスク開始レスポンスを定義する
- [x] 1.3 `frontend/src/e2e/helpers/wails-mock.ts` を拡張し、`MasterPersona` と `ModelSettings` が依存する Wails API を固定データで返せるようにする

## 2. Page Object の拡張

- [x] 2.1 `frontend/src/e2e/page-objects/pages/master-persona.po.ts` に初期表示、NPC 選択、プロンプト確認 / 編集、モデル設定操作、JSON 選択、タスク開始確認の API を追加する
- [x] 2.2 必要に応じて `MasterPersona` 関連コンポーネントへ最小限のテスト識別子や安定 locator を追加し、PageObject からのみ参照する

## 3. Playwright シナリオの実装

- [x] 3.1 `frontend/src/e2e` に `MasterPersona` 必須シナリオ用 spec を追加し、初期表示と NPC 詳細確認のシナリオを実装する
- [x] 3.2 同 spec に `PromptSettingCard` の表示 / 編集境界と `ModelSettings` の主要操作シナリオを実装する
- [x] 3.3 同 spec に JSON 選択からタスク開始状態確認までのシナリオを実装する

## 4. 品質ゲート確認

- [x] 4.1 変更中ファイルに対して `npm run lint:file -- <file...>` を実行し、違反を解消する
- [x] 4.2 `npm run lint:frontend` を実行し、フロントエンド lint 全体が通ることを確認する
- [x] 4.3 Playwright E2E を実行し、`MasterPersona` の必須シナリオが品質ゲートとして通ることを確認する
