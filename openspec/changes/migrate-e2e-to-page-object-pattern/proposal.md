## Why

現状の Playwright E2E は `app.fixture.ts` の単一 harness に操作と検証が集約され、画面追加時に責務が肥大化しやすい。PageObject パターンへ移行して UI 操作責務を分離し、E2E 拡張時の保守性と可読性を改善する必要がある。

## What Changes

- Playwright E2E を `fixture + helper` 中心の構成から、PageObject 中心の構成へ全面移行する
- 画面単位（Dashboard / DictionaryBuilder / MasterPersona）と共通部品単位（Layout / Sidebar / Header）の PageObject を導入する
- spec から直接 locator を扱わない方針に変更し、操作と可視確認を PageObject API に集約する
- 既存 `app.fixture.ts` の harness を廃止し、fixture は Wails mock 初期化と PageObject 注入に限定する
- 既存の最小品質ゲートシナリオ（4ケース）の検証対象は維持しつつ、実装方式のみを置き換える

## Capabilities

### New Capabilities
- `e2e-page-object-architecture`: Playwright E2E における PageObject 分割方針、責務境界、fixture との接続方式を定義する

### Modified Capabilities
- `playwright-quality-gate`: Playwright 品質ゲートの実装方針を、PageObject 前提の構造へ更新する

## Impact

- `frontend/src/e2e` 配下のテスト構成（fixture / helper / spec / 新規 page-objects）
- `frontend/playwright.config.mjs` と E2E 実行導線
- `AGENTS.md` のフロント品質ゲート運用説明（Playwright 実行は維持）
- 将来の E2E シナリオ追加時の実装パターンとレビュー観点
