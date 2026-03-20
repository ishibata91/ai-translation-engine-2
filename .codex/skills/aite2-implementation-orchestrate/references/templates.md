# Implementation Orchestration Templates

## context packet 依頼
```md
### Context Request
- 対象 change:
- 対象 task:
- 読むべき changes:
- 読むべき docs:
- 読むべきコード:
- 今回の焦点:
```

## coder 起動テンプレ
```md
Use $SKILL at skills/$PATH to implement the assigned task.

Task:
- 対象 change:
- 対象 task:
- 担当領域: frontend | backend
- 所有ファイル範囲:

Required reading:
- tasks.md:
- change docs:
- spec:

Execution rules:
- 自分の所有範囲だけを変更する
- 他 agent の変更を巻き戻さない
- 指定品質ゲートを実行して結果を返す

Return format:
### Implementation
- 実装内容:
- 変更ファイル:

### Verification
- 実行コマンド:
- 結果:
- 未検証:
```

## reviewer 起動テンプレ
```md
Use $aite2-implementation-review at skills/aite2-implementation-review to review the integrated implementation diff.

Inputs:
- 対象 change:
- change docs:
- docs/ 正本:
- 実装差分:
- 検証結果:
- 前回 findings:

Return format:
### Findings
- [severity] 根拠つき指摘

### Open Questions
- 追加確認事項

### Residual Risks
- 残留リスク

### Docs Sync
- docs 同期要否
```

## finding 差し戻しテンプレ
```md
Address the following review findings in your owned files first.

Priority:
- 前回 finding の解消確認を最優先にする

Findings:
- [severity] 指摘

Return format:
### Fix Summary
- 解消した finding:
- 解消できない finding:

### Verification
- 再実行コマンド:
- 結果:
```
