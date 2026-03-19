# 辞書DB作成 UI
## MODIFIED Requirements
### Requirement: Dictionary Builder UI Components Integrity
Dictionary Builder UI components MUST preserve existing UX while extracting page logic into a feature hook.
- 既存のUIの見た目 (Tailwind CSS/daisyUI を利用したスタイル) や UX プロセスの流れは維持する。
- 33KB に肥大化したページコンポーネントにおける、ロジック（ステート管理・Wails呼出）を分離する。

#### Scenario: Existing Feature Parity
- **WHEN** ユーザーが辞書構築機能を操作する
- **THEN** 従来と同様に Wails API 呼び出しが行われ、UI が即座に反応する
