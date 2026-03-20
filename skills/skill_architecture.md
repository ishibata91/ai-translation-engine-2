# Skill Architecture

この文書は `skills/` 配下の役割分担と、skill 駆動の orchestration 方針をまとめた構造仕様。
実装済みの内容と、今後この方針へ寄せる内容を分けて扱う。

## 目的

- skill を単機能の手順書として保つ
- `default` を指揮者にして subagent を skill 単位で起動する
- model 指定や profile 指定を skill 本文へ埋め込まない
- design / implementation / bugfix を別フローとして扱い、責務を混ぜない

## Skill と役割

| skill | 主用途 |
| --- | --- |
| `aite2-explorer` | 関連文書、関連コード、ログの圧縮 |
| `aite2-debugger` | 原因予測、デバッグログ配置、再現後の絞り込み |
| `aite2-fixer` | bugfix の限定修正 |
| `aite2-backend-implementation` | backend 実装 |
| `aite2-frontend-implementation` | frontend 実装 |
| `aite2-implementation-review-guard` | 実装差分レビュー |
| `aite2-sync-docs` | docs 正本への同期 |
| `aite2-design-orchestrator` | design 系の入口整理 |

補足:

- profile や model の割り当ては `config.yaml` や外部設定側の責務とする
- skill 本文は profile 名を前提にせず、起動すべき skill 名と責務だけを書く
- 旧来の `context_manager` は skill 名ではなく、今後は `aite2-explorer` に統一する

## 基本原則

- 指揮者 role は `default` が担う
- 1 セクションごとに step-by-step で進める
- `aite2-explorer` は編集せず、圧縮と観測事実の整理だけを行う
- review skill はコード修正を行わない
- design と implementation は自動連結しない
- Playwright ループは skill から言及しない

## Context Board

### 目的

- subagent 間の handoff を口頭説明ではなく、change 配下の共有板で受け渡す
- `aite2-explorer` の圧縮結果、指揮者の判断、review skill / debugger の findings を同じ change 単位で保持する
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
- `aite2-explorer` は board を初期化し、必要な圧縮面を更新する
- review skill / debugger は board に findings を追記し、修正担当はそれを受けて次の手を決める

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
  - review skill / debugger の指摘
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
3. `aite2-explorer` が指示に関連する context を収集し、board を初期化する
4. `aite2-design-orchestrator` が最初に使う設計 skill と順番を決める
5. 各設計 skill を順番に起動する
6. review が必要なら `aite2-design-review-guard` を起動する
7. docs 同期が必要なら `aite2-sync-docs` を起動する

### 設計上の意味

- `aite2-design-orchestrator` は最初に使う design skill と順番を決める司令塔
- `aite2-explorer` は `changes/` `docs/` 既存コードから、設計に必要な範囲だけを圧縮し、次の skill へ渡す context packet を作る
- design 系の change 作成 script は `ui.md` `scenarios.md` `logic.md` に加えて `context_board/` を用意する
- 各 design skill は自分のテンプレを board へ貼り付け、決定事項、未確定事項、次の handoff を残す
- 実際の UI / scenario / logic / review / docs sync の各 skill は、外部 profile ではなく skill 名で分岐して扱う
- context 収集は設計フローの入口で必ず 1 回実施し、途中で不足したときだけ追加で再収集する

### 対象 skill

- `aite2-design-orchestrator`
- `aite2-explorer`
- `aite2-ui-design`
- `aite2-scenario-design`
- `aite2-logic-design`
- `aite2-design-review-guard`
- `aite2-sync-docs`

## Implementation フロー

### 標準チェーン

1. `default` が implementation 依頼を受ける
2. task をセクション単位に分割する
3. `aite2-explorer` が関連文書と関連コードを圧縮する
4. `aite2-backend-implementation` または `aite2-frontend-implementation` が実装する
5. 必要な品質ゲートを実行する
6. `aite2-implementation-review-guard` がレビューする
7. docs 同期が必要な場合だけ `aite2-sync-docs` を使う

### 設計上の意味

- implementation の指揮者は `aite2-implementation-driver`
- mixed task でも v1 は並列化せず、section ごとに順番に処理する
- 実装担当は `aite2-backend-implementation` または `aite2-frontend-implementation` を使う
- review 担当は `aite2-implementation-review-guard` を使い、`critical` と `medium` finding を返す
- docs 同期要否は review 結果と指揮者判断で決め、必要時のみ `aite2-sync-docs` を起動する

### 対象 skill

- `aite2-implementation-driver`
- `aite2-explorer`
- `aite2-backend-implementation`
- `aite2-frontend-implementation`
- `aite2-implementation-review-guard`
- `aite2-sync-docs`

### 実装状態

- impl 系の multi-agent orchestration 基本方針は `aite2-implementation-driver` と実装 / review skill に反映済み
- `aite2-explorer` は未作成だったため追加対象
- frontend skill からは Playwright 前提を外す方針

## Bugfix フロー

### 標準チェーン

1. `default` が bugfix を起動する
2. change 作成 script が `changes/<id>/context_board/` を作る
3. `aite2-explorer` が関連 context を収集し、board を初期化する
4. `aite2-debugger` が原因予測を行い、専用 logger と専用出力を配置する
5. ユーザーが操作してバグを再現する
6. `aite2-log-parser` がログを構造化 JSON にし、起きた事実だけを `aite2-debugger` に返す
7. `aite2-debugger` が原因をさらに絞る
8. bugfix 指揮者が原因を受け取り、実装プランを作る
9. `aite2-fixer` が指示された修正だけを行う
10. `aite2-implementation-review-guard` が review する

### 設計上の意味

- bugfix は implementation とは別の調査駆動フローとする
- `aite2-debugger` は原因予測とデバッグ観測の準備が仕事で、恒久実装は行わない
- 配置する logger は debugger 専用とし、repo 常設ロガーとは分離する
- ログとファイル出力は後で一括削除しやすい配置と命名にする
- `aite2-explorer` は bugfix の入口で context 圧縮と board 初期化を行う
- `aite2-log-parser` は観測事実だけを返し、原因判断は持たない
- `aite2-fixer` は bugfix フローの保守要員として使う
- bugfix 系の change 作成 script も `context_board/` を作り、原因仮説、観測事実、fix plan を board 上で handoff する
- debugger の一時ログと context board は別物として扱う

### 対象 skill

- `aite2-bug-fix`
- `aite2-explorer`
- `aite2-debugger`
- `aite2-log-parser`
- `aite2-fixer`
- `aite2-implementation-review-guard`

### 実装状態

- bugfix 系は旧 role 名の記述が残っていたため、実在 skill 名へ整理する
- `aite2-explorer` `aite2-debugger` `aite2-log-parser` `aite2-fixer` の接続を標準形にそろえる
- 既存 bugfix skill にある `Playwright` 言及は今後除去する

## UI-Refine フロー

### 標準チェーン

1. `default` が ui-refine を起動する
2. change 作成 script が `changes/<id>/context_board/` を作る
3. `aite2-explorer` が対象画面、関連 docs、対象コードを収集し、board を初期化する
4. `aite2-ui-polish` が board を読み、観測、修正方針、変更結果を board に残す
5. 必要なら review skill が見た目修正差分を確認する

### 設計上の意味

- ui-refine は既存の `aite2-ui-polish` を中心に扱う
- 設計変更ではなく既存 UI の見た目修正なので、board は対象画面、観測結果、修正方針の handoff を中心に使う
- ui-refine 系の change 作成 script も `context_board/` を作る

### 対象 skill

- `aite2-explorer`
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

- `aite2-design-orchestrator` を skill 名ベースの接続へ更新
- `aite2-bug-fix` を debugger 主導フローへ更新
- `aite2-explorer` を入口 skill として各フローへ接続する
