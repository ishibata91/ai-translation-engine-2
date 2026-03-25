---
name: plan-direction
description: AI Translation Engine 2 専用。設計依頼、仕様補完、docs 同期のユーザー向け入口であり、plan lane では `plan-distill` 以降の必要な plan chain を自律実行する direction skill。差分仕様、UI 振る舞い、シナリオ、ロジック、設計レビュー、docs 同期までの順序を管理し、自由文の意図が実装や bugfix なら停止して適切な direction skill へ handoff するときにも使う。
---

# Plan Direction

この skill は plan 系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
自分では設計 artifact を作らず、依頼の整理、artifact 充足確認、必要な plan chain の決定と順次実行だけを行う。

## 使う場面
- どの plan skill から始めるべきか迷う
- 差分仕様と実装計画の不足箇所を整理したい
- UI、シナリオ、ロジックのどこを先に確定すべきか決めたい
- 設計レビューや docs 同期まで含めた順番を決めたい

## 入口許可リスト
- ユーザーから直接受けてよいのは設計、仕様補完、docs 同期、artifact 不足整理だけとする。
- `plan-ui` `plan-scenario` `plan-logic` `plan-review` `plan-sync` のような non-direction skill の直指定を受けた場合は、`plan-direction` へ戻す handoff を返す入口として扱う。
- 自由文が実装、UI 反映、コード修正、bugfix、再現、原因調査を要求している場合は、適切な direction skill へ振り分ける conflict 入口として扱う。

## 許可される振り分け
- `plan-direction` の正当入力 lane は `設計 / 仕様補完 / docs 同期 / artifact 不足整理` とする。
- 自由文の意図に `実装 / UI 反映 / task 着手` が含まれる場合は `impl-direction` へ handoff する。
- 自由文の意図に `不具合 / 再現 / 原因切り分け / 修正方針整理` が含まれる場合は `fix-direction` へ handoff する。
- conflict を検出した場合の返答は、`references/templates.md` の conflict template を使った handoff に限る。

## agent / skill 対応
- 判断材料の蒸留は `ctx_loader` に `plan-distill` を使わせる
- UI / シナリオ / ロジックの作成は、それぞれ別々のサブエージェントを起動して `plan-ui` `plan-scenario` `plan-logic` を使わせる
- 設計レビューは `review_cycler` に `plan-review` を使わせる
- docs 同期は `spec_syncer` に `plan-sync` を使わせる

## 手順
1. 依頼が plan lane に属するかを先に判定する。non-direction skill の直指定や impl / fix lane の要求が含まれるなら conflict として停止する。
2. `plan-distill` を起動し、要求、既存 artifact、関連 spec、未確定論点を planning packet に蒸留させる。
3. `plan-distill` を起動した後は packet を待つ。返却前に自分で追加走査、追加読解、下流 skill の先行判断を始めない。
4. planning packet を読んで、依頼が新規設計、設計補完、docs 同期のどれかを分類する。
5. UI、シナリオ、ロジックのうち先に確定すべき層を 1 つ決め、必要な `plan-*` skill の実行順を確定する。
6. 確定した順序に従って必要な `plan-*` skill を自分で順次起動し、各 artifact が埋まるまで chain を進める。
7. `plan-ui` `plan-scenario` `plan-logic` のいずれかを走らせた場合は、artifact が揃った時点で `plan-review` を自動で挟む。
8. `plan-review` の結果を読み、required delta があるか `score < 0.85` の場合は該当する plan skill を再実行して review をやり直す。
9. docs 正本へ昇格すべき差分がある場合だけ `plan-sync` を起動する。
10. review の `score >= 0.85` を満たし、実装へ進める条件が揃ったら `impl-direction` への handoff を明示して終える。

## 標準チェーン
- UI を含む新規設計: `plan-distill` -> `plan-ui` -> `plan-scenario` -> `plan-logic` -> `plan-review` -> `plan-sync`
- UI を含まない責務設計: `plan-distill` -> `plan-scenario` -> `plan-logic` -> `plan-review` -> `plan-sync`
- docs だけを更新したい: `plan-distill` -> `plan-sync`
- 実装準備の artifact 補完: `plan-distill` -> 必要な `plan-*` -> `plan-review` -> `impl-direction`

## 終了条件
- 必要な plan artifact と review 結果が揃い、plan lane で必要な chain が完了している
- 実装へ進む場合は `impl-direction` へ渡す handoff を明示して終える
- conflict の場合は downstream work を始めず、正しい direction skill を明示した handoff を返して終える

## 参照資料
- 振り分け例は `references/routing-examples.md` を読む。
- 順序判断は `references/sequence-checklist.md` を使う。
- `スキルフォルダ/scripts/init-change-design-docs.ps1` の使いどころは `references/doc-init-notes.md` を読む。

## 下流スキル起動時のスキル名明示
- 下流スキルのサブエージェントを立ち上げるときは、必ず `references/templates.md` の「下流スキル起動」テンプレートを使い、`invoked_skill` と `invoked_by` を明示すること
- `invoked_skill` には起動する下流スキル名（例: `plan-distill`）、`invoked_by` には `plan-direction` を設定する
- サブエージェントは起動時にこの情報で「自分がどのスキルとして起動されたか」を確認できる

## 許可される動作
- 指揮役は `orchestration-only` として振る舞い、設計 artifact 作成は下流へ委譲したうえで plan chain の実行管理を担う
- agent 選択は `.codex/agents` を正本として行う
- implementation へ進める判断は、必要な plan artifact が揃った場合に限る
- 各 handoff では確定事項、未確定事項、次の 1 手を明示する
- spec、コード、changes の packet 前読解は `plan-distill` への委譲として扱う
- distill 起動後の次動作は packet 待ちとし、その返却を起点に下流 skill を判断する
- 追加読解が必要な場合は、同じ distill skill の再実行で補う
- plan lane に属する依頼では、必要な plan chain を最後まで自律実行する
- review で required delta が返るか `score < 0.85` の場合は、plan lane の中で回収したうえで impl へ handoff する
- conflict を検出した場合の返答は、適切な direction skill への handoff に限る
