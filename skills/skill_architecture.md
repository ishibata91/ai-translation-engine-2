# Skill Architecture

この文書は `skills/` 配下の役割分担と、profile 駆動の orchestration 方針をまとめた構造仕様。
実装済みの内容と、今後この方針へ寄せる内容を分けて扱う。

## 目的

- skill を単機能の手順書として保つ
- `default` を指揮者にして subagent を役割単位で起動する
- model 指定を skill 本文へ埋め込まず、外部 `config.yaml` profile で差し替えられるようにする
- design / implementation / bugfix を別フローとして扱い、責務を混ぜない

## Profile と役割

| profile | model | sandbox | 主用途 |
| --- | --- | --- | --- |
| `reviewer` | `gpt-5.3-codex` | `read-only` | 実装差分レビュー |
| `tester` | `gpt-5.3-codex` | `workspace-write` | テスト追加、検証、回帰確認 |
| `documenter` | `gpt-5.4-mini` | `workspace-write` | `aite2-sync-docs` による docs 同期 |
| `coder` | `gpt-5.4-mini` | `workspace-write` | 実装、限定的な修正 |
| `architect` | `gpt-5.4` | `workspace-write` | design 系 skill 専用 |
| `context_manager` | `gpt-5.4-mini` | `read-only` | 関連文書、関連コード、ログの圧縮 |
| `debugger` | `gpt-5.4` | `workspace-write` | 原因予測、デバッグログ配置、再現後の絞り込み |

補足:

- `architect` は design 系 skill 専用に固定する
- `context_manger` という typo は使わず、必ず `context_manager` を使う
- model や approval policy は skill 本文に直書きせず、role 名だけを参照する
- role 名は profile 名に合わせて `reviewer` `tester` `documenter` `coder` `architect` `context_manager` `debugger` を使う

## 基本原則

- 指揮者 role は `default` が担う
- 1 セクションごとに step-by-step で進める
- `context_manager` は編集せず、圧縮と観測事実の整理だけを行う
- reviewer は read-only で、コード修正を行わない
- design と implementation は自動連結しない
- Playwright ループは skill から言及しない

## Context Board

### 目的

- subagent 間の handoff を口頭説明ではなく、change 配下の共有板で受け渡す
- `context_manager` の圧縮結果、指揮者の判断、reviewer / debugger の findings を同じ change 単位で保持する
- role ごとの入力契約を固定し、次の skill が何を読むべきかを曖昧にしない

### 配置

- design 系、bugfix 系、ui-refine 系の change 作成 script は `changes/<id>/context_board/` を必ず作る
- `context_board/` は repo 内の共有面だが、自由記述メモ置き場ではなく、テンプレ準拠の handoff 面として扱う
- 共有板は change 配下に閉じ、別 change と混線させない

### 運用ルール

- 各 skill は context 共有用テンプレートを自分の `references/` に持つ
- skill 起動時は、そのテンプレートを `context_board/` 配下へ貼り付けてから中身を書く
- 次の skill は shared board を読んで必要な情報を引き継ぐ
- 指揮者は board のどのファイルを次の role が読むべきかを明示する
- `context_manager` は board を初期化し、必要な圧縮面を更新する
- reviewer / debugger は board に findings を追記し、修正担当はそれを受けて次の手を決める

### 最低限の board 面

- `current_context.md`
  - 対象 task
  - 関連 docs
  - 関連コード
  - 制約
  - 現在の焦点
- `handoff.md`
  - 現在の担当 role
  - 次に起動する skill
  - 完了条件
  - 未確定事項
- `findings.md`
  - reviewer / debugger の指摘
  - 未解消事項
  - 次回の優先論点

補足:

- 実際のファイル名は script 側で調整してよいが、役割はこの 3 面を最低限満たす
- JSON が必要な場合は board 内に別ファイルを追加してよい
- board は削除しやすさより継続的 handoff を優先し、debugger の一時ログとは分離する

## Design フロー

### 標準チェーン

1. `aite2-design-orchestrator` を起動する
2. change 作成 script が `changes/<id>/context_board/` を作る
3. `context_manager` が指示に関連する context を収集し、board を初期化する
4. `aite2-design-orchestrator` が最初に使う設計 skill と順番を決める
5. 各設計 skill を `architect` で担当する
6. review が必要なら `aite2-design-review-guard` も `architect` で担当する
7. docs 同期が必要なら `aite2-sync-docs` も `architect` で担当する

### 設計上の意味

- `aite2-design-orchestrator` は最初に使う design skill と順番を決める司令塔
- `context_manager` は `changes/` `docs/` 既存コードから、設計に必要な範囲だけを圧縮し、architect へ渡す context packet を作る
- design 系の change 作成 script は `ui.md` `scenarios.md` `logic.md` に加えて `context_board/` を用意する
- 各 design skill は自分のテンプレを board へ貼り付け、決定事項、未確定事項、次の handoff を残す
- 実際の UI / scenario / logic / review / docs sync の各 skill は `architect` profile で動く前提にする
- `architect` は implementation や bugfix では使わない
- context 収集は設計フローの入口で必ず 1 回実施し、途中で不足したときだけ追加で再収集する

### 対象 skill

- `aite2-design-orchestrator`
- `aite2-ui-design`
- `aite2-scenario-design`
- `aite2-logic-design`
- `aite2-design-review-guard`
- `aite2-sync-docs`

## Implementation フロー

### 標準チェーン

1. `default` が implementation 依頼を受ける
2. task をセクション単位に分割する
3. `context_manager` が関連文書と関連コードを圧縮する
4. `coder` が実装する
5. `tester` がテスト追加と検証を行う
6. `reviewer` がレビューする
7. docs 同期が必要な場合だけ `documenter` が `aite2-sync-docs` を使う

### 設計上の意味

- implementation の指揮者は `aite2-implementation-driver`
- mixed task でも v1 は並列化せず、section ごとに順番に処理する
- `coder` は実装担当として `aite2-backend-implementation` または `aite2-frontend-implementation` を使う
- `tester` は検証担当として将来専用 skill を持てるように分離する
- `reviewer` は `aite2-implementation-review-guard` を使い、`critical` と `medium` finding を返す
- docs 同期要否は reviewer が判断し、必要時のみ `documenter` を起動する

### 対象 skill

- `aite2-implementation-driver`
- `aite2-backend-implementation`
- `aite2-frontend-implementation`
- `aite2-implementation-review-guard`
- `aite2-sync-docs`

### 実装状態

- impl 系の multi-agent orchestration 基本方針は `aite2-implementation-driver` と coder / reviewer skill に反映済み
- `tester` 専用 skill と `context_manager` 専用 skill は未作成
- frontend skill からは Playwright 前提を外す方針

## Bugfix フロー

### 標準チェーン

1. `default` が bugfix を起動する
2. change 作成 script が `changes/<id>/context_board/` を作る
3. `context_manager` が関連 context を収集し、board を初期化する
4. `debugger` が原因予測を行い、専用 logger と専用出力を配置する
5. ユーザーが操作してバグを再現する
6. `context_manager` role の `aite2-log-parser` がログを構造化 JSON にし、起きた事実だけを `debugger` に返す
7. `debugger` が原因をさらに絞る
8. bugfix 指揮者が原因を受け取り、実装プランを作る
9. `coder` role の `aite2-fixer` が指示された修正だけを行う
10. `reviewer` が review する

### 設計上の意味

- bugfix は implementation とは別の調査駆動フローとする
- `debugger` は原因予測とデバッグ観測の準備が仕事で、恒久実装は行わない
- 配置する logger は debugger 専用とし、repo 常設ロガーとは分離する
- ログとファイル出力は後で一括削除しやすい配置と命名にする
- `context_manager` は `aite2-log-parser` を通じたログ事実化を含む read-only 役として扱う
- `coder` は bugfix フローでは `aite2-fixer` を通じて保守要員として使う
- bugfix 系の change 作成 script も `context_board/` を作り、原因仮説、観測事実、fix plan を board 上で handoff する
- debugger の一時ログと context board は別物として扱う

### 対象 skill

- `aite2-bug-fix`
- `aite2-debugger`
- `aite2-log-parser`
- `aite2-fixer`
- `aite2-implementation-review-guard`

### 実装状態

- bugfix 系はまだ旧 skill 構造のまま
- `debugger` `aite2-log-parser` `aite2-fixer` は追加済みだが、呼び出し元 skill の接続は今後調整余地がある
- 既存 bugfix skill にある `Playwright` 言及は今後除去する

## UI-Refine フロー

### 標準チェーン

1. `default` が ui-refine を起動する
2. change 作成 script が `changes/<id>/context_board/` を作る
3. `context_manager` が対象画面、関連 docs、対象コードを収集し、board を初期化する
4. `aite2-ui-polish` が board を読み、観測、修正方針、変更結果を board に残す
5. 必要なら reviewer が見た目修正差分を確認する

### 設計上の意味

- ui-refine は既存の `aite2-ui-polish` を中心に扱う
- 設計変更ではなく既存 UI の見た目修正なので、board は対象画面、観測結果、修正方針の handoff を中心に使う
- ui-refine 系の change 作成 script も `context_board/` を作る

### 対象 skill

- `aite2-ui-polish`

## Skill の責務境界

### 指揮者 skill

- 役割分解
- 次に使う skill の決定
- subagent 起動
- agent 間の受け渡し
- 終了判定

### 実務 skill

- 自分の role に対応する task だけを処理する
- 他 role の判断を上書きしない
- 必要な入力契約と出力契約を持つ

### review / sync skill

- `aite2-implementation-review-guard` は finding を返すだけで修正しない
- `aite2-sync-docs` は docs 正本への昇格だけを担当し、実装判断はしない

## 命名と構成ルール

- skill 名は `aite2-<role-or-domain>` を基本形にする
- `SKILL.md` に workflow の中核を書く
- 長い prompt 骨子やテンプレは `references/` に逃がす
- `agents/openai.yaml` は UI 表示名、説明、default prompt を持つ
- role ごとの model 差し替えは `config.yaml` profile 側で行う

## 今後の更新対象

- `aite2-design-orchestrator` を `architect` 前提へ更新
- `aite2-bug-fix` を debugger 主導フローへ更新
- `tester` と `context_manager` の専用 skill を導入するかを再検討
