## MODIFIED Requirements

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
- **THEN** そのページの品質ゲートは `e2e-required-scenarios` に定義された `必須シナリオ` がすべて成功した場合にのみ合格しなければならない

#### Scenario: Regression checks are implemented through page object APIs
- **WHEN** 開発者または AI が Playwright 品質ゲートの spec を追加または修正する
- **THEN** 画面操作と可視確認は PageObject API を介して実装されなければならない
- **AND** spec 側で `page.getBy*` や `page.locator` を直接増やしてはならない
