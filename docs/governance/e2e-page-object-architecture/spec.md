# E2E ページオブジェクト設計

## Purpose

Playwright E2E を PageObject 中心の責務分離構造へ統一し、spec の可読性と保守性を担保する。

## Requirements

### Requirement: Playwright E2E MUST Use PageObject-Centric Test Structure
システムは Playwright E2E の実装構造を PageObject 中心へ統一しなければならない。spec ファイルはシナリオ手順の記述に専念し、UI locator の直接参照を標準形としてはならない。

#### Scenario: Spec does not contain direct locator operations
- **WHEN** 開発者が `frontend/src/e2e` 配下の spec を追加または修正する
- **THEN** spec は PageObject API を呼び出す構成でなければならない
- **AND** `page.getBy*` や `page.locator` の直接利用を常態化してはならない

### Requirement: PageObjects MUST Be Split By Shared Components and Page-Specific Behaviors
システムは PageObject を「共通部品」と「画面固有」の単位で分割しなければならない。共通 UI（例: Layout / Sidebar / Header）と画面固有 UI（例: Dashboard / DictionaryBuilder / MasterPersona）を別責務として保持しなければならない。

#### Scenario: Shared shell logic is reused across scenarios
- **WHEN** 複数シナリオで同じナビゲーションや共通レイアウト確認を行う
- **THEN** 共通部品 PageObject の API を再利用しなければならない
- **AND** 同一 locator や同一操作を各画面 PageObject に重複定義してはならない

### Requirement: Fixture MUST Initialize Environment and Inject PageObjects Only
システムは E2E fixture の責務を環境初期化と PageObject 注入に限定しなければならない。Wails mock 初期化は fixture で行ってよいが、画面固有操作や画面固有検証を fixture 内に持ってはならない。

#### Scenario: Fixture provides PageObject instances
- **WHEN** テスト実行時に fixture が起動する
- **THEN** fixture は Wails mock を初期化したうえで PageObject インスタンス群を提供しなければならない
- **AND** 旧来の単一 harness へ操作/検証を集約する実装へ戻してはならない

### Requirement: Runtime Error Detection MUST Be Preserved During PageObject Migration
システムは PageObject 移行後も、画面遷移と初期表示の各タイミングで runtime 例外検知を維持しなければならない。

#### Scenario: Page transition triggers runtime exception
- **WHEN** 画面遷移後に Wails binding 不足や初期化失敗が発生する
- **THEN** E2E は可視要素待ちでタイムアウトする前に runtime 例外として失敗理由を報告しなければならない
