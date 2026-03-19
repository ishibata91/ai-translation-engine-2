# Review

## Findings

### 1. `translationinput` に Terminology 対象抽出責務を持たせる案は artifact 境界を越えやすい
- `logic.md` では `translationinput` が parser 出力から Terminology 対象を抽出し、正規化済み `[]TerminologyEntry` を task 単位で保存・返却する正本になるとしている。
- 一方で architecture では artifact は handoff 保存境界に留まり、業務判断や進行決定を持たないことが求められている。
- 「どの REC を Terminology 対象にするか」は terminology 側の業務判断に近く、artifact に固定すると後の変更で保存境界まで巻き込みやすい。

参照:
- [logic.md](/F:/ai translation engine 2/changes/align-terminology-targets-with-dictionary-import/logic.md)
- [spec.md](/F:/ai translation engine 2/docs/governance/architecture/spec.md)

### 2. `logic.md` の NPC request 構築ルールがまだ曖昧
- `RecordType + SourceText` は非 NPC の重複統合キーとしては成立する。
- ただし NPC は `FULL/SHRT` を同時翻訳する仕様のため、request 構築単位は `PairKey` を優先する必要がある。
- 例えば同じ `NPC_:FULL` と同じ `SourceText` を持つ別 NPC が異なる `PairKey` / `SHRT` を持つ場合、`RecordType + SourceText` だけではどの short 名と組にするか判別できない。
- したがって finding は「`RecordType + SourceText` 方針自体が誤り」ではなく、「NPC 例外を本文で明示する必要がある」という内容に留めるべきである。

参照:
- [logic.md](/F:/ai translation engine 2/changes/align-terminology-targets-with-dictionary-import/logic.md)
- [spec.md](/F:/ai translation engine 2/docs/slice/terminology/spec.md)

### 3. Terminology spec の独自 DTO 方針と、change の `translationinput` 直読方針が未同期
- Terminology spec は、本 slice 独自 DTO を受け取り、他 package DTO へ依存しない方針を持っている。
- change は `translationinput` の正規化 DTO を terminology がそのまま読む前提になっている。
- このまま実装すると `docs/slice/terminology/spec.md` との不一致が残る。

参照:
- [logic.md](/F:/ai translation engine 2/changes/align-terminology-targets-with-dictionary-import/logic.md)
- [spec.md](/F:/ai translation engine 2/docs/slice/terminology/spec.md)

## Open Questions
- `translationinput` は Terminology 専用の正規化配列まで保存するべきか。それとも parser row の保存に留め、terminology slice で投影するべきか。
- NPC 以外の重複統合は `RecordType + SourceText` で固定してよいか。
- terminology spec の独自 DTO 方針を維持するか、`translationinput` DTO 直読へ方針変更するか。

## Residual Risks
- `docs/slice/dictionary/spec.md` と `docs/slice/terminology/spec.md` はまだ change と同期されていない。
- `changes/align-terminology-targets-with-dictionary-import/` には `scenarios.md` と `tasks.md` が無く、実装へ落とす時の確認観点が不足している。

## Notes
- 追加確認で、NPC の懸念は `FULL` と `SHRT` の REC type 衝突ではなく、同じ `RecordType + SourceText` を持つ別 `PairKey` 同士が 1 request に潰れる点だと整理した。
- そのため、NPC は request 構築キーに `PairKey` を優先し、非 NPC だけ `RecordType + SourceText` を重複統合キーに使う案が必要になる。
