# Delta Specs
## MODIFIED Requirements
### Requirement: Master Persona UI Components Integrity
Master Persona UI components MUST preserve existing UX while extracting page logic into a feature hook.
- 既存のUIの見た目 (Tailwind CSS/daisyUI を利用したスタイル) や UX プロセスの流れは維持する。
- 51KB に肥大化したページコンポーネントにおける、ロジック（ステート管理・Wails呼出）を分離する。

#### Scenario: Existing Feature Parity
- **WHEN** ユーザーがパーソナ生成の開始、一時停止、ログの閲覧など既存の操作を行う
- **THEN** 従来と同様に Wails API 呼び出しが行われ、UI が即座に反応する
