# Logic Design

## Scenario
翻訳フローの terminology phase で、対象語 preview に翻訳済みテキストを合成し、実行中 progress を workflow public contract として公開する。

## Goal
`単語翻訳` phase の UI が、対象語一覧・翻訳済みテキスト・進捗状態を単一の workflow 契約として取得できるようにし、frontend が slice 内部の保存構造や runtime の生イベントを直接解釈せずに済むようにする。

## Responsibility Split
- Controller:
  - Wails から `ListTranslationFlowTerminologyTargets`、`RunTranslationFlowTerminology`、`GetTranslationFlowTerminology` を公開する。
  - request / response の境界整形だけを行い、進捗意味付けや translated text の結合は行わない。
- Workflow:
  - translation-flow store から terminology 対象一覧を読み、terminology slice から翻訳済み結果と phase summary / progress snapshot を取得する。
  - page 単位の対象語 row に対して `translated_text` と `translation_state` を合成し、frontend 用 DTO を返す。
  - terminology phase の public status を `pending` `running` `completed` `completed_partial` `run_error` に固定する。
  - progress public contract を `progress_mode` `progress_current` `progress_total` `progress_message` に固定する。
  - 実行中 progress は foundation progress bridge の `translation_flow.terminology.progress` トピックへ公開し、`GetTranslationFlowTerminology` も同じ snapshot shape を返す。
  - 実行開始中の `refresh` / `pagination` / `advance` 可否を phase state として判断する。
- Slice:
  - `translationflow` slice は terminology 対象入力の保存・ページング・task 境界管理を維持する。
  - `terminology` slice は prompt 準備、結果保存、phase summary 算出に加え、preview row ID を key にした translated text lookup を提供する。
  - duplicate 統合や NPC FULL/SHRT の解釈、保存済み訳の代表値決定と preview row への fan-out は terminology slice が担う。
- Artifact:
  - 新しい shared artifact は追加しない。
  - terminology phase の一覧表示に必要な入力集合は既存の task 境界に保存し、他 phase 共有が必要になった場合のみ artifact 昇格を再検討する。
- Runtime:
  - terminology 実行の外部 I/O を実行し、進捗イベントを workflow へ返す。
  - 進捗イベントは `current`、`total`、`message` のような transport 事実に留め、phase 意味付けは持たない。
- Gateway:
  - LLM 実行や外部接続の技術的依頼口を提供する。
  - translated text の UI 表示や phase 進行判定には関与しない。

## Data Flow
- 入力:
  - `task_id`
  - terminology target preview の `page` / `page_size`
  - terminology 実行用 `TranslationRequestConfig` / `TranslationPromptConfig`
- 中間成果物:
  - translation-flow store が返す terminology target entries
  - terminology slice が返す phase summary
  - terminology slice が返す preview row ID 単位の translated text lookup
  - runtime が返す progress event / completion result
- 出力:
  - `TerminologyTargetPreviewPage.rows[]` に `translated_text` と `translation_state` を含む row 群
  - `TerminologyPhaseResult` に `status` `saved_count` `failed_count` `progress_mode` `progress_current` `progress_total` `progress_message` を含む phase 状態

## Main Path
1. Controller が terminology 対象一覧取得要求を workflow へ渡す。
2. Workflow が translation-flow store から対象 row page を取得する。
3. Workflow が同ページの preview row IDs を terminology slice へ渡し、保存済み translated text lookup を取得する。
4. Workflow が row ごとに `translated_text` と `translation_state` を合成し、frontend DTO として返す。
5. Controller が terminology 実行要求を workflow へ渡す。
6. Workflow が terminology slice に prompt 準備を依頼し、runtime executor に LLM 実行を委譲する。
7. Runtime が進捗イベントを返すたび、workflow が terminology phase progress snapshot を更新し、progress bridge と `GetTranslationFlowTerminology` の双方で同じ shape を公開する。
8. Workflow が terminology slice に結果保存を依頼し、保存後に summary と translated text lookup が同じ task 境界から再取得可能な状態にする。
9. Workflow が terminal summary を `completed` または `completed_partial` で返し、後続の `GetTranslationFlowTerminology` / `ListTranslationFlowTerminologyTargets` で同じ結果を再構築できるようにする。

## Key Branches
- target 一覧取得時に保存済み訳がない:
  - workflow は row を `translation_state=missing` として返し、frontend は `未翻訳` 表示にする。
- runtime が総数を返せない:
  - workflow は `progress_mode=indeterminate` を返し、frontend は不定 progress を表示する。
- terminology 実行が部分成功で終わる:
  - terminology slice は成功分だけ保存し、summary は `status=completed_partial` と `failed_count > 0` を返す。
  - workflow は失敗行だけ `translation_state=missing` のまま返す。
- terminology 実行全体が失敗する:
  - workflow は `status=run_error` と `progress_mode=hidden` を返し、最後に確定していた translated text lookup は保持したまま返す。
- resume / retry / cancel:
  - resume は `GetTranslationFlowTerminology` と対象一覧再取得で最後の確定状態を復元する。
  - retry は `run_error` から同じ task 境界で再度 `RunTranslationFlowTerminology` を許可する。
  - `completed_partial` 後は同 phase 内の再実行導線を追加せず、failure を可視化したまま後続 phase 進行を許可する。
  - cancel は本 change では追加せず、workflow も専用 cancel API を持たない。

## Persistence Boundary
- translation-flow store:
  - terminology 対象入力、ページング元データ、task と source file の対応を保持する。
- terminology slice local persistence:
  - 保存済み translated text
  - phase summary (`status`, `saved_count`, `failed_count`)
  - preview row ID を key にした translated text lookup
- runtime / workflow state:
  - 実行中 progress snapshot は実行状態として保持する。
  - terminal 後に永続化が必要なのは summary と saved translations であり、進捗の逐次イベント列までは永続化しない。

## Side Effects
- runtime executor が LLM 実行を行う。
- terminology slice が保存済み訳と summary を更新する。
- workflow が foundation progress bridge の `translation_flow.terminology.progress` へ進行事実を通知する。
- frontend は実行中に progress bridge を購読し、phase 初期化・再表示・再読込では `GetTranslationFlowTerminology` を呼んで最新 snapshot を復元する。

## Risks
- 現行 `TerminologyTargetPreviewRow` に `translated_text` が無いため、workflow DTO・controller binding・frontend adapter を同時更新する必要がある。
- 現行 `RunTranslationFlowTerminology` は同期完了型のため、progress bridge の配線が不足していると UI は terminal result しか受け取れない。
- preview row ID を terminology slice の fan-out key に使うため、既存保存データに row ID 対応が無い場合は read-model 補完または migration が必要になる。

## Context Board Entry
```md
### Logic Design Handoff
- 確定した責務境界: workflow が preview row ID 単位で saved translation lookup を合成し、status と progress snapshot を workflow public contract として公開する
- docs 昇格候補: terminology target preview row の `translated_text` / `translation_state` 契約、phase progress contract、`completed_partial` と `run_error` の terminal ルール
- review で見たい論点: progress bridge shape と `GetTranslationFlowTerminology` の整合、preview row ID fan-out の成立条件
- 未確定事項: なし
```
