# Impl/Fix Skill運用ハードニング

## 概要

`impl-direction` 配下で発生した停止・再開・reroute の手戻りを減らし、同系統の問題を `fix-direction` 配下でも先回りして抑えるための差分仕様。

今回の change は、コード生成能力そのものではなく、以下の運用基盤を対象にする。

- 進捗の正本管理
- condensed packet による context 圧縮
- worker / fixer の停止条件と blocked 契約
- orchestrator が持つ一時資産のライフサイクル管理

## 対象

- `impl-direction`
- `impl-workplan`
- `impl-backend-work`
- `impl-frontend-work`
- `fix-direction`
- `fix-distill`
- `fix-trace`
- `fix-analysis`
- `fix-work`
- `fix-review`
- `fix-logging`

## Artifacts

- [UI](/F:/ai translation engine 2/changes/impl-fix-skill-ops-hardening/ui.md)
- [Scenarios](/F:/ai translation engine 2/changes/impl-fix-skill-ops-hardening/scenarios.md)
- [Logic](/F:/ai translation engine 2/changes/impl-fix-skill-ops-hardening/logic.md)
- [Tasks](/F:/ai translation engine 2/changes/impl-fix-skill-ops-hardening/tasks.md)
