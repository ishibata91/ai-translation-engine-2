---
name: investigation-direction
description: AI Translation Engine 2 専用。調査タスクのオーケストレーター。コードベースの広範な調査、特定機能の実装箇所の探索、仕様の確認などを指揮する。調査は下流の skill に委譲し、自身は調査フローの管理のみを行う。
---

# Investigation Direction

この skill は調査（Investigation）系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
下流 skill に調査対象の特定（distill）と実際のコード探索（explorer）を委譲し、その手順と結果の集約を管理する。

## agent / skill 対応
- 調査対象（ファイル、パッケージ、シンボル）の特定とリスト化は `ctx_loader` に `investigation-distill` を使わせる
- 特定されたポインタ群からの実際のコード探索や詳細調査は `investigator` に `investigation-explorer` を使わせる

## 入口許可リスト
- ユーザーから直接受けてよいのは、仕様調査、実装箇所の探索、特定の仕組みの解明など、コードベースに対する「調査」に関するものだけとする。
- `investigation-distill` や `investigation-explorer` のような non-direction skill の直指定を受けた場合は、`investigation-direction` へ戻す handoff を返す入口として扱う。
- 自由文が不具合修正（fix）、通常実装（impl）、設計・仕様変更（plan）を明確に要求している場合は、適切な direction skill（`fix-direction`, `impl-direction`, `plan-direction`）へ振り分ける conflict 入口として扱う。

## 許可される振り分け
- `investigation-direction` の正当入力 lane は `コード調査 / 仕様の確認 / 実装箇所の特定` とする。
- 自由文の意図に `実装変更 / 新規タスク` が含まれる場合は `impl-direction` へ handoff する。
- 自由文の意図に `不具合の修正 / 再現確認` が含まれる場合は `fix-direction` へ handoff する。
- conflict を検出した場合の返答は、`references/templates.md` の conflict template を使った handoff に限る。

## 許可される運用範囲
- 指揮役の責務は orchestration に限る。
- 調査結果を元に実装へ進む場合は、対応する direction skill への handoff として扱う。
- handoff はパケットとして直接引き渡す。

## やること
1. 依頼が investigation lane に属するかを先に判定する。
2. `investigation-distill` を起動し、調査対象に関連しそうなファイル、パッケージ、シンボルのポインタをリスト化させる。
3. `investigation-distill` からリストを受け取るまで待機する。自身でコード探索を行わない。
4. 返却されたポインタリストを元に `investigation-explorer` を起動し、実際のコード走査と詳細な調査を行わせる。
5. `investigation-explorer` からの調査結果を受け取り、ユーザーに報告用に整形して回答する。
6. 追加の調査が必要な場合は、再度 `investigation-distill` または `investigation-explorer` を呼び出す。

## 参照
- 記録テンプレートは `references/templates.md` を使う。

## 下流スキル起動時のスキル名明示
- 下流スキルのサブエージェントを立ち上げるときは、必ず `references/templates.md` の「下流スキル起動」テンプレートを使い、`invoked_skill` と `invoked_by` を明示すること。
- `invoked_skill` には起動する下流スキル名（例: `investigation-distill`）、`invoked_by` には `investigation-direction` を設定する。
- サブエージェントは起動時にこの情報で「自分がどのスキルとして起動されたか」を確認できる。

## 許可される動作
- 指揮役は `orchestration-only` として振る舞い、調査対象の蒸留と詳細探索の接続に注力する。
- distill / explorer を適切に呼び出し、得られた結果を繋げることに注力する。
