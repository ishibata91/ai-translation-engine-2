# Backend Implementation Templates

## 実装メモ
```md
### Implementation
- 対象 section:
- section_goal:
- 関連設計文書:
- 所有ファイル範囲:
- 変更対象ファイル:
- 実装方針:
```

## 確認メモ
```md
### Verification
- backend:lint:file:
- lint:backend:
- validation_result:
- noise_classification:
- 未検証:
- 未完了:
```

## section 結果
```md
### Section Result
- section_id:
- result: completed | blocked
- completed_scope:
- remaining_gap:
- changed_paths:
- validation_result:
- noise_classification: none | scope_failure | external_validation_noise | known_pre_existing_issue
- reroute_hint:
- unverified:
```

## Blocked 返却メモ
```md
### Blocked
- 対象 section:
- 停止理由:
- completed_scope:
- remaining_gap:
- noise_classification:
- owned_paths 外で必要になった対象:
- 該当する別 section または contract:
- reroute_hint:
- impl-direction へ返す次アクション:
```

## 修正返却メモ
```md
### Fix Summary
- 対象 section:
- 解消した finding:
- 解消できない finding:

### Verification
- 再実行コマンド:
- 結果:
```
