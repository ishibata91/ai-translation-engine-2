## Why

`MasterPersona` は現在、サイドバーから画面へ遷移できることしか Playwright で担保されておらず、主要な操作導線や結果状態の退行を検出できていない。特に `ModelSettings` と `PromptSettingCard` はページ内の重要な設定 UI であるにもかかわらず E2E で守られていないため、`e2e-required-scenarios` の運用に合わせてこのページ全体を品質ゲート対象として扱える状態にする必要がある。

## What Changes

- `MasterPersona` ページに対するページ単位 E2E の必須シナリオを定義する
- 必須シナリオに対応する Playwright E2E を追加し、初期表示だけでなく NPC 一覧 / 詳細導線、`ModelSettings` の主要設定操作、`PromptSettingCard` の表示 / 編集導線を検証できるようにする
- 既存の Page Object 構造を維持しつつ、`MasterPersona` 用 page object と fixture/mock の責務を必要最小限で拡張する

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `e2e-required-scenarios`: `MasterPersona` ページ向けの必須シナリオ定義を追加し、ページ品質ゲートとして通すべき操作と到達状態を明文化する

## Impact

- OpenSpec: `openspec/specs/e2e-required-scenarios` 配下に `MasterPersona` 向けページ spec を追加または拡張する
- Frontend E2E: `frontend/src/e2e` 配下の `MasterPersona` 関連 spec、page object、fixture/mock データに影響する
- UI 対象: `frontend/src/pages/MasterPersona.tsx` に加え、`frontend/src/components/ModelSettings.tsx` と `frontend/src/components/masterPersona/PromptSettingCard.tsx` の主要導線が E2E 対象に含まれる
- 品質ゲート: `npm run lint:frontend` と Playwright E2E の確認対象に `MasterPersona` の必須シナリオが加わる
- API / DB / 外部依存: 変更なし。既存の Playwright と現在のフロントエンド構成を前提に進める
