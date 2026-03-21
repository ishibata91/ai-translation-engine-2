# Tasks

## 1. Normalized Target Planning
- [ ] terminology slice に raw input 正規化の共通 planner を追加し、日本語行・空文字・非対象 REC を preview / execute の両方で同じ規則で除外する
- [ ] workflow の terminology preview 取得を raw artifact 直ページングから切り替え、normalized target set を表示できるようにする
- [ ] NPC FULL / SHRT と duplicate 統合後の target_count が phase summary と一致するように整理する

## 2. Dictionary Artifact Resolution
- [ ] terminology slice で dictionary_artifact を正本として全文完全一致を検索し、exact match を `cached` 結果へ変換する
- [ ] exact 未解決 group に対して `original_source_text` を保持したまま `replaced_source_text` を生成できる request 契約を追加する
- [ ] keyword exact partial replacement 用の matcher に strict keyword boundary / longest-first / non-overlap を保証させ、`Skeever Den -> スキーヴァー Den` 型の部分置換を決定できるようにする
- [ ] partial replacement 用の exact keyword 検索と reference_terms 用検索を分離し、stemming / LIKE ベースの候補を部分置換へ流用しないようにする
- [ ] unresolved group に対して `replaced_source_text` を前提に keyword exact replacement 後の reference terms を付与する
- [ ] NPC group では人名部分一致候補を reference-only で追加し、強制置換には使わない

## 3. Phase Execution and Persistence
- [ ] `PreparePrompts` 後に cached 保存済み group 数を summary に反映し、未解決 request が 0 件なら runtime なしで完了できるようにする
- [ ] prompt 構築で LLM に渡す本文を `source_text` 直使用から `replaced_source_text` 正本へ切り替え、保存キーは `original_source_text` を維持する
- [ ] retry / resume 時に cached group を再 dispatch せず、未解決 group に同じ partial replacement を再適用して再実行する
- [ ] preview translation 復元で cached / translated / missing の各状態を一貫して返す

## 4. Verification and Docs
- [ ] terminology slice / workflow テストで、日本語除外、exact match short-circuit、NPC 部分一致、preview と execute の同一 target set を確認する
- [ ] `Skeever Den` での部分置換、重複区間競合時の長い一致優先、strict keyword boundary / non-overlap、retry 再適用を検証ケースへ追加する
- [ ] terminology phase の partial completion / run_error でも cached 結果が保持されることを確認する
- [ ] `docs/slice/terminology/spec.md` と `docs/slice/terminology/terminology_test_spec.md` を current contract と今回の規則へ同期する
