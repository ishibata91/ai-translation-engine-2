## MODIFIED Requirements

### Requirement: architecture.md は構造責務だけを保持しなければならない
`architecture.md` は、責務区分、依存方向、DTO / Contract / DI 原則、composition root の責務だけを保持しなければならない。品質ゲート、テスト設計、ログ運用、フロント構造の詳細を内包してはならず、これらの共通基準は `governance` 区分の専用 spec へ委譲しなければならない。

#### Scenario: 品質ゲートやログ規約が別 spec へ委譲される
- **WHEN** 開発者が品質ゲート、テスト設計、ログ運用、フロント構造を確認したい
- **THEN** `architecture.md` は専用 spec を参照するだけに留まらなければならない
- **AND** 具体的な運用ルールは `governance/backend-quality-gates/spec.md`、`governance/standard-test/spec.md`、`governance/log-guide/spec.md`、`frontend/frontend-architecture/spec.md` に置かれなければならない

### Requirement: 共通要件の配置先は責務ごとに固定されなければならない
複数ユースケースで共有される要件は、以下の配置基準に従って共通 spec へ配置しなければならない。

- `governance` は architecture、spec-structure、quality-gates、test standard、log standard、repo-wide requirements を扱う
- `frontend` は React 画面、ページ構造、UI レイアウト、画面遷移、フロント専用設計を扱う
- `controller` は Wails binding、HTTP、CLI など外部入力の受け口契約を扱う
- `workflow` は phase 管理、resume / cancel、進行制御、DTO マッピングを扱う
- `slice` は個別ユースケース固有の振る舞い、DTO、契約、補助図を扱う
- `runtime` は queue、task、executor、進行制御基盤を扱う
- `artifact` は slice 間 handoff の保存・検索境界を扱う
- `gateway` は外部資源への依頼口と技術接続を扱う
- `foundation` は telemetry、progress などの横断基盤を扱う

#### Scenario: ユースケース spec に共通要件を書こうとする
- **WHEN** 開発者が UI / workflow / runtime / artifact / gateway / foundation の共通ルールをユースケース spec に追加しようとする
- **THEN** システムはその要件を責務に応じた共通 spec へ移す前提で整理しなければならない
- **AND** ユースケース spec には、そのユースケース固有の振る舞いと契約だけを残さなければならない

### Requirement: spec の分類は実装責務に最も近い区分へ合わせなければならない
spec の置き場は、機能名や画面名ではなく、最終的にどの責務区分の判断材料として使う文書かで決めなければならない。文書のタイトルや起点 UI が同じでも、主題が UI なら `frontend`、入口契約なら `controller`、進行制御なら `workflow`、共有 handoff なら `artifact`、横断通知基盤なら `foundation` を選ばなければならない。

#### Scenario: Wails 画面の見た目ではなく Go binding 契約を定義する
- **WHEN** 文書が Wails binding の公開メソッドや request/response 契約を扱う
- **THEN** その文書は `frontend` ではなく `controller` 区分へ置かなければならない

#### Scenario: UI から開始されるが本質は phase 進行を定義する
- **WHEN** 文書がボタン押下後の enqueue / dispatch / save / resume の進行規則を扱う
- **THEN** その文書は `frontend` や `controller` ではなく `workflow` または `runtime` 区分へ置かなければならない

#### Scenario: 見た目は画面機能だが中身は UI 表示責務である
- **WHEN** 文書が画面表示、view model、入力 UI、レイアウト、page hook 境界を扱う
- **THEN** その文書は feature 名にかかわらず `frontend` 区分へ置かなければならない
- **AND** `slice` 区分へ置いてはならない

### Requirement: runtime / gateway / UI の共通 spec はユースケース spec と分離されなければならない
`queue`、`task`、`artifact`、`progress`、`telemetry`、`config`、`datastore`、`frontend-headless-architecture`、`wails-app-shell` などの基盤 spec は、特定ユースケースに閉じない共通機能として扱わなければならない。

#### Scenario: 基盤要件が特定ユースケース spec に埋め込まれている
- **WHEN** resume、progress、queue、artifact handoff、gateway、Wails binding などの共通要件がユースケース spec へ追加される
- **THEN** システムはそれを共通 spec へ移し、ユースケース spec からは参照で接続しなければならない

### Requirement: AGENTS.md は spec 選択の入口を示さなければならない
`AGENTS.md` は、設計・提案・実装時に参照すべき spec を責務ごとに案内しなければならない。AI が `architecture.md` に品質ルールやログ規約を書き戻さない構成でなければならず、canonical path は `openspec/specs/<zone>/<capability>/...` を前提に示さなければならない。

#### Scenario: AI が文書責務に応じて参照先を決める
- **WHEN** AI がアーキテクチャ、品質ゲート、テスト設計、ログ設計、spec 配置方針を検討する
- **THEN** `AGENTS.md` から該当 spec を一意に辿れなければならない
- **AND** root 直下の旧パスを前提にした案内を残してはならない

## ADDED Requirements

### Requirement: spec の物理配置は zone/capability の canonical path へ統一されなければならない
システムは、spec の物理配置を `openspec/specs/<zone>/<capability>/spec.md` またはその配下の補助文書へ統一しなければならない。`openspec/specs` root 直下へ単独の `.md` 文書を正本として置いてはならない。

#### Scenario: 新しい capability spec を追加する
- **WHEN** 開発者が新しい capability spec を追加する
- **THEN** 文書は `openspec/specs/<zone>/<capability>/spec.md` に配置されなければならない
- **AND** root 直下へ `<name>.md` を追加してはならない

#### Scenario: 補助文書を同居させる
- **WHEN** ある capability が test scope、interface note、supplement を持つ
- **THEN** 補助文書は同じ capability ディレクトリ配下へ配置されなければならない
- **AND** 他 zone へ分散してはならない

### Requirement: 誤分類または混在 spec は移設または分割されなければならない
システムは、spec の内容と配置区分が一致しない場合、正しい zone へ移設しなければならない。1 つの spec に複数責務区分が明確に混在する場合は、責務ごとに分割しなければならない。

#### Scenario: UI spec が slice 配下に置かれている
- **WHEN** `master-persona-ui` のように UI 表示責務を持つ spec が `slice` 配下に存在する
- **THEN** 当該 spec は `frontend` 区分へ移設されなければならない
- **AND** `slice` 配下へ残してはならない

#### Scenario: UI と workflow が 1 文書へ混在している
- **WHEN** `translation-flow-data-load` のように入力 UI と phase 進行の両方を 1 文書で扱っている
- **THEN** システムは `frontend` と `workflow` の別 capability へ分割しなければならない
- **AND** 元文書に混在状態を残してはならない
