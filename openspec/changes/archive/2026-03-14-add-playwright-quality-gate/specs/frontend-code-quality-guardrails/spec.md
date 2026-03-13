## MODIFIED Requirements

### Requirement: Fast Lint For Files In Progress
フロントエンド開発フローは、全体 lint とは別に、開発者が編集中のファイルのみを対象として素早く実行できる lint 導線を提供しなければならない。部分実行の結果は、全体 lint と同じ規約セットのうち対象ファイルに適用可能な検査を実行し、TSDoc と公開範囲最小化の違反を確認できなければならず、この要件を MUST とする。さらに、変更完了前の品質ゲートは `lint:file` と `lint:frontend` で局所・全体の lint を確認した後に、Playwright E2E による統合退行確認へ進む標準フローを持たなければならない。

#### Scenario: Run Lint On Current File
- **WHEN** 開発者が編集中の 1 ファイルを指定してフロントエンド lint を実行する
- **THEN** そのファイルに適用される規約だけが高速に検査され、違反があれば失敗しなければならない

#### Scenario: Run Full Lint Before Completion
- **WHEN** 開発者が作業完了前にフロントエンド全体の lint を実行する
- **THEN** 部分実行では検出できない横断的な違反も含めて、最終的な品質ゲートとして評価されなければならない

#### Scenario: Run Playwright after frontend lint passes
- **WHEN** 開発者または AI がフロントエンド変更の完了判定を行う
- **THEN** `lint:file` と `lint:frontend` の通過後に Playwright E2E を実行しなければならない

### Requirement: AI-Driven Incremental Lint Remediation
フロントエンドの変更フローは、変更対象ファイルに対する機械的な lint 実行結果を AI が逐次解釈し、規約違反の修正に即時利用できなければならない。lint の出力は対象ファイル、違反ルール、位置、修正に必要な要点を安定して取得できる形式でなければならず、この要件を MUST とする。加えて、AI はフロントエンド変更時に `lint:file -> 修正 -> 再実行 -> lint:frontend -> Playwright` の順で品質確認を進めなければならない。

#### Scenario: AI Reads File-Scoped Lint Result
- **WHEN** 開発者または自動化フローが変更中ファイルを対象に lint を実行する
- **THEN** AI は出力結果から対象ファイルと違反内容を特定し、その場で修正案を適用できなければならない

#### Scenario: Rule Not Fully Lintable
- **WHEN** `frontend_coding_standards.md` の規約のうち静的解析だけでは完全に判定できない項目が存在する
- **THEN** 開発フローは、その項目を AI の逐次確認対象として扱い、機械検査で代替できない補完チェックとして明示しなければならない

#### Scenario: AI executes E2E as the final frontend gate
- **WHEN** AI がフロントエンド変更の最終確認を実施する
- **THEN** `lint:frontend` の後段で Playwright E2E を品質ゲートとして実行しなければならない
