# Design Review

score: 0.91

### Design Review Findings
- なし

### Open Questions
- なし

### Residual Risks
- preview と execute が別経路で target 集合を作る実装にすると、日本語除外や target_count が再びずれる余地がある
- exact match を `PreparePrompts` で保存する設計を採る場合、retry 時に cached 済み group を再送しないガードが必要

### Docs Sync
- 要否: 必要。`docs/slice/terminology/spec.md` と `docs/slice/terminology/terminology_test_spec.md` へ昇格候補を反映する
