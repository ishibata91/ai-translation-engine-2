# Implementation Readiness and Routing Checklist

## artifact readiness
- `ui.md` `scenarios.md` `logic.md` の主張が矛盾していない
- 実装判断を止める unknowns が plan artifact 側に残っていない
- `tasks.md` は前提 artifact ではなく、`impl-workplan` が生成する出力 artifact として扱う
- 正本 chain は `impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review`
- `Workplan Summary` の各 section は `title / owner / goal / depends_on / shared_contract / required_reading / owned_paths / forbidden_paths / validation_commands / acceptance` をすべて保持する

## frontend section のシグナル
- 主変更が `frontend/src` 配下
- `ui.md` または frontend spec が主判断材料
- 品質ゲートが `typecheck / lint:frontend / Playwright`

## backend section のシグナル
- 主変更が `pkg/` や controller / workflow / slice / runtime / artifact / gateway
- `logic.md` や architecture spec が主判断材料
- 品質ゲートが `backend:lint:file / lint:backend`

## routing matrix の原則
- `ui.md` が無いことだけを理由に backend-only と判定しない
- frontend 影響有無は `owned_paths` `goal` `validation_commands` `acceptance` を含む section signal 全体で判定する
- `frontend/src` 変更や frontend 品質ゲートが含まれるなら、`ui.md` 不在でも frontend section 候補として扱う
- section はモジュール/契約単位で分ける
- 1 section = 1 owner を守る
- shared contract は `impl-workplan` で固定してから worker を起動する
- 1 つの worker に両方の品質ゲートを背負わせない
- review reroute は `affected_sections` 単位で行い、元の full work order 契約を崩さない
- reroute payload でも `shared_contract / required_reading / validation_commands / acceptance` を落とさない
