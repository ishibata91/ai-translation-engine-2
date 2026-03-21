# Logic Design

## Scenario
翻訳フローの terminology phase で、raw terminology input を日本語除外込みで正規化し、マスター辞書 artifact の全文完全一致を即時保存しつつ、全文未解決行にはキーワード単位完全一致の部分置換を適用してから reference terms を付与し、LLM 実行する。

## Goal
terminology phase が、Dictionary Builder の shared artifact を唯一の辞書正本として使い、日本語行の誤送信を防ぎながら、全文 exact match は deterministic に確定し、keyword exact match は `cached` に昇格させず LLM 入力本文の前処理として適用し、NPC の部分一致は reference-only で補助する責務分担を成立させること。

## Responsibility Split
- Controller:
  - `ListTranslationFlowTerminologyTargets`、`RunTranslationFlowTerminology`、`GetTranslationFlowTerminology` の DTO 境界を維持する。
  - 日本語判定、辞書検索、exact match 判定、keyword exact partial replacement、reference_terms の組み立ては持たない。
- Workflow:
  - task 境界の raw terminology input を読み出し、preview と execute の両方で terminology slice の正規化結果を使う。
  - phase summary の起点を「LLM request 数」ではなく「日本語除外後の翻訳単位総数」に合わせて管理する。
  - `PreparePrompts` 後に pending request が 0 件でも cached 保存済み単位があれば、runtime を呼ばずに phase を完了扱いにする。
  - retry / resume では、同じ正規化ルール、同じ keyword replacement ルール、同じ cached 結果を前提に未解決分だけを再 dispatch する。
- Slice:
  - terminology slice が raw input の正規化、日本語除外、duplicate / NPC ペア維持、全文完全一致、keyword exact partial replacement、reference_terms 抽出、prompt 生成を一貫して担う。
  - exact match の辞書訳は `cached` として terminology store へ保存し、LLM dispatch 対象から除外する。
  - exact 未解決 target には `original_source_text` を保持したまま `replaced_source_text` を生成し、LLM へ渡す本文は常に `replaced_source_text` を正本とする。
  - keyword exact partial replacement は terminology slice 内の matcher が strict keyword boundary、longest-first、non-overlap を保証し、曖昧ヒットや stemming ヒットを使わない。
  - searcher は部分置換用の exact keyword 候補集合と reference_terms 用候補集合を分けて返し、matcher が部分置換に使う候補を決定する。
  - NPC の人名部分一致は reference-only とし、確定置換へ昇格させない。
  - preview 用の対象集合も execute 用の対象集合も同じ内部 planner から導出する。
- Artifact:
  - `translationinput` artifact は raw terminology input の正本を保持する。
  - `dictionary_artifact` はマスター辞書 source/entry の正本を保持し、全文 exact / keyword exact / reference_terms / NPC partial の検索元になる。
  - artifact は対象除外、keyword replacement、exact match 採否を判断しない。
- Runtime:
  - terminology slice が返した未解決 request だけを LLM へ dispatch する。
  - cached 保存済みの行や日本語除外行の存在を前提に、`replaced_source_text` を含む dispatch 対象 request 数だけを実行する。
- Gateway:
  - dictionary_artifact repository と LLM 実行の技術接続を提供する。
  - 日本語除外、exact match 採否、keyword replacement、NPC 部分一致の業務ルールは持たない。

## Data Flow
- 入力:
  - `task_id`
  - task 境界の raw `translationinput.TerminologyInput`
  - `dictionary_artifact.Entry`
  - terminology phase の request / prompt config
- 中間成果物:
  - 日本語除外後の normalized target groups
  - exact match で `cached` 保存される deterministic result 群
  - unresolved groups ごとの `original_source_text`
  - unresolved groups ごとの `replaced_source_text`
  - keyword exact replacement に使う match spans
  - unresolved groups に対する `reference_terms`
  - unresolved groups のみから作られる `[]llmio.Request`
- 出力:
  - preview 用 terminology target page
  - `cached` と `translated` を含む preview translations
  - phase summary (`target_count`, `saved_count`, `failed_count`, progress snapshot)
  - terminology store に保存された訳語

## Main Path
1. Workflow が `translationinput` artifact から raw terminology input を取得する。
2. Workflow が terminology slice に raw input を渡し、shared REC allow-list と日本語除外を含む normalized target groups を構築させる。
3. terminology slice が各 group に対してマスター辞書 artifact の全文完全一致を検索する。
4. exact match が見つかった group は terminology slice が `cached` 結果へ変換し、terminology store に保存する。
5. exact 未解決 group に対して terminology slice が keyword exact 候補を取得し、matcher が strict keyword boundary・longest-first・non-overlap で置換区間を確定する。
6. terminology slice が `original_source_text` を保持したまま `replaced_source_text` を生成し、部分置換は `cached` にせず unresolved group の属性として保持する。
7. terminology slice が `replaced_source_text` を前提に reference_terms を構築し、NPC group では人名部分一致候補を追加取得する。
8. terminology slice が unresolved group のみで prompt を構築し、`replaced_source_text` と `reference_terms` を埋めた `[]llmio.Request` を返す。
9. Workflow が `target_count = normalized groups`, `saved_count = cached groups`, `progress_current = cached groups` で phase summary を更新する。
10. Workflow が unresolved request だけを runtime へ dispatch し、レスポンスを terminology slice `SaveResults` へ渡す。
11. terminology slice がレスポンスを保存し、cached + translated の合算結果を preview / summary から再構築可能な状態にする。

## Key Branches
- 日本語行:
  - raw input に対象 REC が含まれても、日本語判定に一致した行は normalized target groups に入れない。
  - preview / execute / retry の全経路で同じ除外ルールを適用する。
- exact match のみで完了:
  - unresolved request が 0 件で cached group が 1 件以上ある場合、workflow は runtime を呼ばず `completed` を返す。
  - preview は cached 訳を翻訳済みとして表示する。
- keyword exact partial replacement:
  - exact 未解決 group に keyword exact 候補がある場合だけ、terminology slice は `replaced_source_text` を生成する。
  - 部分置換は strict keyword boundary の一致だけを対象にし、候補競合時は longest-first / non-overlap で決定する。
  - 部分置換で全文が日本語化しても `cached` へ昇格させず、LLM request は unresolved として維持する。
- NPC 部分一致:
  - NPC group に全文完全一致が無い場合だけ、人名部分一致候補を `reference_terms` に追加する。
  - 人名部分一致候補は keyword exact replacement の対象外で、prompt に補助情報として共存させる。
- partial completion / run_error:
  - cached 保存済み group は terminal 状態でも保持される。
  - runtime 失敗やレスポンス欠損があっても、未解決 group だけが `missing` のまま残り、`replaced_source_text` 自体は永続化対象にしない。
- resume / retry / cancel:
  - resume は raw input を再ロードしつつ、同じ normalized target planner、同じ keyword replacement planner、同じ保存済み訳を使って現在状態を復元する。
  - retry は cached group を再 dispatch せず、未解決 group だけに同じ partial replacement を再適用して再実行する。
  - cancel は本 change では追加しない。

## Persistence Boundary
- `translationinput` artifact:
  - raw terminology input、source file、preview row ID、NPC pair の基礎情報を保持する。
  - 日本語除外後の派生 target set や `replaced_source_text` は正本として保存しない。
- `dictionary_artifact`:
  - source 単位の辞書 entry を shared artifact 正本として保持する。
  - terminology phase はここから exact / keyword exact / reference_terms / NPC partial を検索する。
- terminology store:
  - `cached` を含む訳語保存結果
  - phase summary
  - preview row に対応する translated / missing 状態
- runtime / workflow state:
  - unresolved request の進捗 snapshot だけを一時的に保持する。
  - 日本語除外行の逐次イベント、match spans、`replaced_source_text`、reference_terms 自体は永続化しない。

## Side Effects
- terminology slice が dictionary_artifact を検索する。
- terminology slice が exact match を terminology store へ事前保存する。
- terminology slice が keyword exact match spans を評価し、`replaced_source_text` を生成する。
- workflow / runtime が unresolved request だけを LLM 実行する。
- terminology slice が LLM 結果を terminology store へ保存する。
- workflow が phase summary と progress を更新する。

## Risks
- preview と execute の両方で同じ normalized target planner を使わないと、日本語除外や target_count が UI と実行で不整合になる。
- exact match を `PreparePrompts` 時点で保存するため、retry 時に cached 済み group を重複 dispatch しない phase summary 制御が必要になる。
- keyword exact replacement に `searcher.go` の stemming + LIKE を流用すると strict keyword boundary を満たせないため、reference_terms 用検索と責務を分ける必要がある。
- `original_source_text` と `replaced_source_text` を request 契約で分離しないと、保存キーと LLM 入力本文が混線する。
- 日本語判定を NPC FULL / SHRT ペアへ適用する順序を誤ると、片側だけ日本語のペア処理が不整合になる。
- `docs/slice/terminology/terminology_test_spec.md` は `PreparePrompts` / `SaveResults` 契約と keyword exact partial replacement の検証ケースへ更新済みであり、実装時は `replaced_source_text` と retry 再適用の観点をその正本に追従しないと挙動差分を生みやすい。

## Context Board Entry
```md
### Logic Design Handoff
- 確定した責務境界: terminology slice が raw input 正規化・日本語除外・exact match・keyword exact partial replacement・reference_terms 構築を担い、workflow は normalized target と cached 進捗を前提に phase を進行する
- docs 昇格候補: 日本語行は preview/execute 両方で除外する規則、exact match を cached として事前保存する規則、keyword exact replacement は strict keyword boundary / longest-first / non-overlap で `replaced_source_text` を作る規則、NPC 部分一致は reference-only とする規則
- review で見たい論点: preview と execute の target set 一貫性、request 契約の `original_source_text` / `replaced_source_text` 分離、reference_terms 用検索と keyword replacement 用検索の責務分離、partial replacement 観点の test spec 同期要否
- 未確定事項: なし
```
