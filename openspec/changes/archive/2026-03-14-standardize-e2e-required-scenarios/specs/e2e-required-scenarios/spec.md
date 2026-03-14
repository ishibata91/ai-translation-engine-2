## ADDED Requirements

### Requirement: Page-level E2E required scenarios SHALL define pass conditions in a page-agnostic way
システムは、ページ単位 E2E の合格条件を画面固有名称ではなく `必須シナリオ` の概念で定義しなければならない。`必須シナリオ` は、そのページにおける 1 つのユーザー目的を完結させる最小の統合シナリオでなければならない。

#### Scenario: Required scenario is defined as a complete user goal
- **WHEN** 開発者があるページの E2E 品質ゲートを設計する
- **THEN** 各 `必須シナリオ` は、そのページで利用者が達成したい 1 つの目的を開始から完了まで通して検証する単位として定義されなければならない

### Requirement: Page-level E2E required scenarios MUST be split into a standard spec and page-specific specs
システムは、ページ単位 E2E の共通ルールを親 spec に、各ページ固有の必須シナリオを配下ディレクトリの spec に分離して管理しなければならない。親 spec は標準定義だけを扱い、ページ固有のシナリオ列挙を直接抱えてはならない。

#### Scenario: Standard rules and page-specific rules are stored separately
- **WHEN** 開発者が `e2e-required-scenarios` capability を追加または更新する
- **THEN** 共通用語、シナリオ粒度、合格条件は `specs/e2e-required-scenarios/spec.md` に記述されなければならない
- **AND** 各ページ固有の必須シナリオは `specs/e2e-required-scenarios/<page>/spec.md` に記述されなければならない

### Requirement: Page-level E2E gates SHALL pass only when all required scenarios pass
システムは、あるページのページ単位 E2E を品質ゲートとして扱う場合、そのページに定義された `必須シナリオ` がすべて成功したときにのみ合格とみなさなければならない。

#### Scenario: One required scenario fails
- **WHEN** あるページに定義された `必須シナリオ` のうち 1 つでも失敗する
- **THEN** そのページの E2E 品質ゲートは不合格として扱われなければならない

#### Scenario: All required scenarios succeed
- **WHEN** あるページに定義された `必須シナリオ` がすべて成功する
- **THEN** そのページの E2E 品質ゲートは合格として扱われなければならない

### Requirement: Required scenarios MUST detect rendering, interaction, and transition regressions
システムは、ページ単位 E2E の `必須シナリオ` により、少なくとも初期描画、主要操作、結果画面または状態遷移の退行を検出できなければならない。

#### Scenario: Page only verifies initial rendering
- **WHEN** あるページの E2E が初期表示の可視確認だけで終了している
- **THEN** そのシナリオは `必須シナリオ` としては不十分であり、主要操作と到達結果を含む形に拡張されなければならない

#### Scenario: Required scenario covers user interaction to destination state
- **WHEN** あるページの `必須シナリオ` が定義される
- **THEN** そのシナリオは初期描画確認に加えて、主要操作の実行と、操作後に到達すべき画面または状態の確認を含まなければならない
