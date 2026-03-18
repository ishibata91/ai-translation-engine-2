# Spec: playwright-quality-gate

## Overview

Playwright を用いたフロントエンド E2E テスト基盤と、最低限の統合退行を検出する品質ゲート要件を定義する。

## Requirements

### Requirement: Frontend Workspace SHALL Provide Playwright E2E Foundation
システムは `frontend` ワークスペース内に Playwright ベースの E2E テスト基盤を提供しなければならない。基盤には Playwright 設定、実行 script、共通 fixture または helper、および初期テスト配置規約を含めなければならない。

#### Scenario: Frontend workspace can run Playwright
- **WHEN** 開発者が `frontend` ワークスペースで Playwright 用 script を実行する
- **THEN** Playwright 設定が読み込まれ、E2E テストディレクトリを対象にブラウザテストを開始できなければならない

#### Scenario: Shared test bootstrap is reusable
- **WHEN** 開発者が新しい E2E シナリオを追加する
- **THEN** ベース URL、初期待機、主要レイアウト確認などの共通処理を fixture または helper から再利用できなければならない

### Requirement: Playwright E2E SHALL Verify Minimum UI Shell Regression Gate
システムは Playwright E2E により、少なくともアプリ起動、主要レイアウト表示、主要画面導線の退行を検出できなければならない。初期品質ゲートは `Dashboard`、`DictionaryBuilder`、`MasterPersona` への基本遷移を対象にしなければならない。さらに、ページ単位 E2E を品質ゲートへ組み込む場合は `e2e-required-scenarios` capability に定義された `必須シナリオ` を通過条件に含めなければならない。品質ゲート実装は PageObject 中心の構造で維持されなければならず、spec 側で直接 locator 操作を拡散させてはならない。

#### Scenario: App shell is visible after startup
- **WHEN** Playwright がアプリのベース URL を開く
- **THEN** ヘッダー、サイドバー、メインコンテンツ領域などの主要レイアウトが表示されなければならない

#### Scenario: Dashboard is rendered as the default route
- **WHEN** Playwright が初期表示直後の画面を検証する
- **THEN** ダッシュボードの初期表示を識別できる要素が存在しなければならない

#### Scenario: Major pages are reachable from the shell
- **WHEN** Playwright が主要導線を操作する
- **THEN** `DictionaryBuilder` と `MasterPersona` のページへ遷移し、それぞれの画面を識別できなければならない

#### Scenario: Page-level E2E gate requires all required scenarios to pass
- **WHEN** あるページを Playwright 品質ゲートの対象として運用する
- **THEN** そのページの品質ゲートは、`e2e-required-scenarios` に定義された `必須シナリオ` がすべて成功した場合にのみ合格とみなされなければならない
- **AND** いずれかの `必須シナリオ` が失敗した場合は品質ゲートを不合格にしなければならない

#### Scenario: Regression checks are implemented through page object APIs
- **WHEN** 開発者または AI が Playwright 品質ゲートの spec を追加または修正する
- **THEN** 画面操作と可視確認は PageObject API を介して実装されなければならない
- **AND** spec 側で `page.getBy*` や `page.locator` を直接増やしてはならない

### Requirement: MVP Playwright Scope MUST Prefer Browser E2E Over Native Wails Automation
システムは Playwright 基盤の MVP において、Wails ネイティブウィンドウの直接自動化ではなく、Vite で起動した frontend ブラウザアプリを検証対象にしなければならない。Wails 固有ランタイムの完全統合検証は、この capability の必須要件に含めてはならない。

#### Scenario: Browser-based execution is used for the initial gate
- **WHEN** Playwright 品質ゲートの実行方式を確認する
- **THEN** 実行対象は `frontend` のブラウザアプリでなければならない

#### Scenario: Native Wails automation is not required for MVP pass condition
- **WHEN** 開発者が Playwright 基盤の初期実装を完了する
- **THEN** Wails デスクトップバイナリの直接操作を実装しなくても MVP 要件を満たせなければならない
