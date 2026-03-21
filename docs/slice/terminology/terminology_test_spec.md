# 用語翻訳・用語DB保存 テスト仕様

本設計は `docs/governance/standard-test/spec.md` と `docs/governance/log-guide/spec.md` に準拠し、`pkg/slice/terminology` の公開 Contract を境界にした Table-Driven Test を定義する。細粒度ヘルパの個別ユニットテストではなく、phase 単位の入力と保存結果を一貫して検証する。

## 1. テスト方針

1. **Contract 起点**: `PreparePrompts(ctx, taskID, options)`、`SaveResults(ctx, taskID, responses)`、`GetPreviewTranslations(ctx, entries)`、`GetPhaseSummary(ctx, taskID)` を検証境界とする。
2. **Table-Driven Test**: raw terminology input、dictionary_artifact の初期状態、モック LLM 応答、期待 summary / preview / 保存結果を表形式で与える。
3. **構造化ログ前提**: 各サブテストは一意の TraceID を持つ `context.Context` を引き回し、失敗時は構造化ログで差分を確認できる状態にする。
4. **2フェーズ分離の検証**: `PreparePrompts` は request 生成・cached 解決・partial replacement 本文生成、`SaveResults` は LLM 応答保存を検証対象として分ける。

## 2. パラメタライズドテストケース

### 2.1 `Terminology.PreparePrompts` テスト

| ケースID | 目的 | 初期状態 | アクション | 期待結果 |
| :------- | :--- | :------- | :--------- | :------- |
| TTP-01 | 日本語行を terminology target から除外する | raw input に対象 REC だが日本語の `SourceText` を含む | `PreparePrompts(ctx, taskID, options)` | 日本語行は request 数、exact match 検索、preview target のいずれにも含まれない |
| TTP-02 | 完全一致を cached として short-circuit する | dictionary_artifact に `Auriel's Bow -> アーリエルの弓` の完全一致がある | `PreparePrompts(ctx, taskID, options)` | 当該行の LLM request は生成されず、store には `status=cached` の保存結果が残る |
| TTP-03 | NPC 人名部分一致を reference-only で付与する | dictionary_artifact に `Jon Battle-Born` 等の NPC entry があり、入力は全文完全一致しない NPC 行 | `PreparePrompts(ctx, taskID, options)` | unresolved request が 1 件生成され、prompt の `reference_terms` に NPC 部分一致候補が含まれる |
| TTP-04 | 非 NPC 重複を 1 つの翻訳単位へ統合する | 同じ `RecordType + SourceText` を持つ複数行が存在する | `PreparePrompts(ctx, taskID, options)` | request 数は 1 件になり、summary の target_count も統合後単位で数えられる |
| TTP-05 | exact match のみで phase 完了可能な状態を作る | 正規化後の全 group が dictionary_artifact 完全一致で解決できる | `PreparePrompts(ctx, taskID, options)` | request 数は 0 件、cached 件数は target_count と一致し、workflow が runtime なしで完了扱いにできる |
| TTP-06 | `Skeever Den` の keyword exact partial replacement を生成する | dictionary_artifact に `Skeever -> スキーヴァー` があり、対象行は全文完全一致しない | `PreparePrompts(ctx, taskID, options)` | unresolved request は維持され、LLM 入力本文は `スキーヴァー Den` になり、`cached` へ昇格しない |
| TTP-07 | 重複区間競合で longest-first / non-overlap を守る | `Broken Tower Redoubt` と `Broken` が同時に一致候補になる | `PreparePrompts(ctx, taskID, options)` | 置換区間は長い候補が優先され、短い候補が重複区間を上書きしない |
| TTP-08 | `original_source_text` と `replaced_source_text` の契約分離を守る | keyword exact partial replacement が成立する未解決行がある | `PreparePrompts(ctx, taskID, options)` | 保存キーに使う原文は維持され、LLM 入力本文だけが `replaced_source_text` へ切り替わる |

### 2.2 `Terminology.SaveResults` テスト

| ケースID | 目的 | 初期状態 | アクション | 期待結果 |
| :------- | :--- | :------- | :--------- | :------- |
| TTS-01 | unresolved request の正常保存 | `PreparePrompts` 済みで unresolved request がある | `SaveResults(ctx, taskID, responses)` | `TL: |...|` を抽出して保存し、preview では `translated` として復元される |
| TTS-02 | NPC FULL / SHRT の fan-out 保存 | paired NPC request に対する 1 件の LLM 応答がある | `SaveResults(ctx, taskID, responses)` | FULL / SHRT それぞれの record_type で保存され、preview から両行の訳が復元される |
| TTS-03 | 一部失敗時も cached / 成功済み結果を保持する | cached 済み group と unresolved group が混在し、応答の一部が失敗する | `SaveResults(ctx, taskID, responses)` | cached 結果と成功応答は保持され、失敗分だけ `missing` のまま残る |
| TTS-04 | 形式不正レスポンスを安全にスキップする | `TL: |...|` を含まない応答が渡される | `SaveResults(ctx, taskID, responses)` | 当該 group は保存されず、他の正常応答だけが保存される |

### 2.3 Workflow / Preview 整合テスト

| ケースID | 目的 | 初期状態 | アクション | 期待結果 |
| :------- | :--- | :------- | :--------- | :------- |
| TTW-01 | preview と execute が同じ target set を使う | raw input に日本語行、exact match 行、unresolved 行が混在する | `ListTerminologyTargets` -> `RunTerminologyPhase` | preview に現れる対象集合と execution 対象集合が一致し、日本語行だけが両方から除外される |
| TTW-02 | 既存 task 再表示で cached / translated が復元される | terminology 実行済み task を持つ | `GetTerminologyPhase` と preview 取得 | exact match と LLM 翻訳の両方が保存済み結果として復元される |
| TTW-03 | retry 時に cached 済み group を再送しない | partial completion または run_error 後の task | 再度 `RunTerminologyPhase` | cached 済み group は request に含まれず、未解決 group だけが再 dispatch される |
| TTW-04 | retry 時に partial replacement ルールを再適用する | 初回実行で `replaced_source_text` を使った未解決 group がある | 再度 `RunTerminologyPhase` | 未解決 group に同じ strict keyword boundary / longest-first / non-overlap ルールが再適用される |

## 3. ログとデバッグ

### 3.1 テスト基盤でのログ準備
- サブテストごとに一意の TraceID を持つ `context.Context` を生成する
- `slog` 出力はテスト単位でファイルへ記録し、phase summary 更新・dictionary 検索・cached 保存・partial replacement 生成・LLM 保存を追跡できるようにする

### 3.2 AI デバッグプロンプト

```text
以下は terminology slice のテスト実行ログである。
`docs/slice/terminology/spec.md` と `docs/slice/terminology/terminology_test_spec.md` の期待動作と比較し、
日本語除外、cached exact match、keyword exact partial replacement（`replaced_source_text`）、NPC partial reference、preview/execute/retry 整合のどこで乖離したかを特定して修正案を示せ。

--- 実行ログ ---
{ログファイル内容}

--- 期待仕様 ---
{該当セクション}
```
