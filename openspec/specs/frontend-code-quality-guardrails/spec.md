# Spec: frontend-code-quality-guardrails

## Overview

フロントエンドのコーディング規約を、可能な限り lint と補助的な静的検査で機械検証し、変更中ファイルに対する AI の逐次修正フローへ接続する。

## Requirements

### Requirement: TSDoc Coverage For Public Frontend Contracts
フロントエンドの TypeScript / React コードにおいて、他ファイルから参照される主要な公開契約は TSDoc により目的と利用境界を説明しなければならない。lint は `frontend/src` 配下の公開コンポーネント、公開 Hook、公開型、公開ユーティリティ関数のうち、規約で必須とされた対象に TSDoc が存在しない場合に失敗しなければならず、この要件を MUST とする。

#### Scenario: Missing TSDoc On Exported Hook
- **WHEN** 開発者が `frontend/src` 配下で公開 Hook を `export` し、規約上必須の TSDoc を付けずに lint を実行する
- **THEN** lint は失敗し、対象シンボルに TSDoc が不足していることを示さなければならない

#### Scenario: TSDoc Present On Public Contract
- **WHEN** 開発者が規約対象の公開コンポーネントまたは公開型に TSDoc を記述して lint を実行する
- **THEN** lint は TSDoc 要件を満たしたものとして通過しなければならない

### Requirement: Minimize File-Level Public Surface
フロントエンドの TypeScript ファイルは、ファイル外から利用される識別子のみを公開しなければならない。lint または補助的な静的検査は、未使用の `export`、または同一ファイル内に閉じ込められる実装詳細の公開を検出し、不要な公開範囲の縮小を促さなければならず、この要件を MUST とする。

#### Scenario: Internal Helper Is Exported
- **WHEN** 開発者が feature hook 内部でのみ使う補助関数や補助型を `export` した状態で検査を実行する
- **THEN** 検査は、その公開が不要であることを報告しなければならない

#### Scenario: Required Public API Remains Exported
- **WHEN** ページや他モジュールから参照される公開 Hook や公開型だけを `export` した状態で検査を実行する
- **THEN** 検査は、必要な公開契約を維持したまま通過しなければならない

### Requirement: Fast Lint For Files In Progress
フロントエンド開発フローは、全体 lint とは別に、開発者が編集中のファイルのみを対象として素早く実行できる lint 導線を提供しなければならない。部分実行の結果は、全体 lint と同じ規約セットのうち対象ファイルに適用可能な検査を実行し、TSDoc と公開範囲最小化の違反を確認できなければならず、この要件を MUST とする。

#### Scenario: Run Lint On Current File
- **WHEN** 開発者が編集中の 1 ファイルを指定してフロントエンド lint を実行する
- **THEN** そのファイルに適用される規約だけが高速に検査され、違反があれば失敗しなければならない

#### Scenario: Run Full Lint Before Completion
- **WHEN** 開発者が作業完了前にフロントエンド全体の lint を実行する
- **THEN** 部分実行では検出できない横断的な違反も含めて、最終的な品質ゲートとして評価されなければならない

### Requirement: AI-Driven Incremental Lint Remediation
フロントエンドの変更フローは、変更対象ファイルに対する機械的な lint 実行結果を AI が逐次解釈し、規約違反の修正に即時利用できなければならない。lint の出力は対象ファイル、違反ルール、位置、修正に必要な要点を安定して取得できる形式でなければならず、この要件を MUST とする。

#### Scenario: AI Reads File-Scoped Lint Result
- **WHEN** 開発者または自動化フローが変更中ファイルを対象に lint を実行する
- **THEN** AI は出力結果から対象ファイルと違反内容を特定し、その場で修正案を適用できなければならない

#### Scenario: Rule Not Fully Lintable
- **WHEN** `frontend_coding_standards.md` の規約のうち静的解析だけでは完全に判定できない項目が存在する
- **THEN** 開発フローは、その項目を AI の逐次確認対象として扱い、機械検査で代替できない補完チェックとして明示しなければならない

### Requirement: Maximize Rule Coverage From Frontend Coding Standards
`frontend_coding_standards.md` に定義された規約のうち、静的解析で判定可能なものは可能な限り lint ルールへ変換しなければならない。完全な lint 化が難しい規約については、どの規約を lint で担保し、どの規約を AI の逐次確認で補完するかを開発フロー上で識別可能にしなければならず、この要件を MUST とする。

#### Scenario: Rule Can Be Enforced By Lint
- **WHEN** 規約項目が import 境界、TSDoc、不要 export、`any` 禁止のように静的解析で判定可能である
- **THEN** その規約は AI の注意喚起だけで済ませず、lint ルールとして実装されなければならない

#### Scenario: Rule Requires AI Judgment
- **WHEN** 規約項目が責務の適切な分離や読解負荷の高さのように静的解析だけでは十分に判定できない
- **THEN** その規約は AI が変更ごとに確認する補完ルールとして扱われ、lint 対象外であることが明示されなければならない
