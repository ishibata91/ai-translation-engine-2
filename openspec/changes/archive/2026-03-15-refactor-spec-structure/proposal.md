## Why

現在の `openspec/specs` は、`architecture.md` が定義する責務区分と物理配置が一致しておらず、旧配置と新配置も混在している。特に `cross-cutting` に寄せた分類は `artifact` と `foundation` の独立境界を弱め、`master-persona-ui` や `translation-flow-data-load` のように UI / workflow / slice の責務が混ざった spec を生みやすくしているため、参照導線と配置ルールを正規化する必要がある。

## What Changes

- `openspec/specs` の正規分類を `governance / frontend / controller / workflow / slice / runtime / artifact / gateway / foundation` の 9 区分へ整理する。
- `spec-structure` を更新し、正規配置、旧配置からの移行ルール、補助文書の同居ルール、root 直下 `.md` 廃止ルールを定義する。
- `architecture` を更新し、spec 側の分類軸が実装責務区分と一致するよう参照先と用語を整理する。
- 旧配置と新配置が混在している spec を正規パスへ移動し、`master-persona-ui` のような誤分類 spec を frontend へ戻す。
- `translation-flow-data-load` のように UI と workflow 要件が混在する spec は責務ごとに分割し、元 spec には混在要件を残さない。
- `AGENTS.md` の参照導線を新しい spec 構造に合わせて更新する。

## Capabilities

### New Capabilities
- `governance`: OpenSpec 文書群の共通基準、品質ゲート、配置ルール、全体要求を集約する正規区分を定義する
- `frontend/translation-flow-data-load-ui`: 翻訳フローのデータロード画面に属する UI 要件を frontend 区分で定義する
- `workflow/translation-flow-data-load`: 翻訳フローのデータロード工程に属する phase 進行と受け渡し要件を workflow 区分で定義する

### Modified Capabilities
- `spec-structure`: spec の正規分類、物理配置、移行ルール、混在 spec の分割基準を変更する
- `architecture`: spec 参照導線を `artifact` と `foundation` を含む実装責務区分に揃える
- `translation-flow-data-load`: 混在していた UI / workflow 要件を新しい frontend / workflow capability へ分離し、残すべき責務だけに絞る

## Impact

- OpenSpec: `openspec/specs` 配下のディレクトリ構成、複数 spec の物理パス、相互参照、`spec-structure` / `architecture` / `translation-flow-data-load` の要件
- 運用導線: `AGENTS.md` から辿る設計・品質・テスト・ログ・frontend/backend 参照先
- 実装影響: なし。対象は文書構造と OpenSpec artifact 群のみ
- CLI 影響: capability パス変更後も `openspec/specs/<zone>/<capability>/spec.md` として辿れるよう、参照切れを残さない
