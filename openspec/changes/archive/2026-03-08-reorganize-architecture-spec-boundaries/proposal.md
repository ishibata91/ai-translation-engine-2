## Why

現状の `architecture.md` とユースケース spec 群は、アーキテクチャ説明、UI 要件、runtime 要件、品質ルールが混在しており、どの文書を更新すべきか判断しにくい。実装 change を小さく保つためには、まず仕様書の責務境界を整理し、共通構造をユースケース spec から切り離す必要がある。

## What Changes

- `architecture.md` を純粋なアーキテクチャ文書として再構成する
- spec フォルダ全体を棚卸しし、UI / workflow / runtime / gateway の共通要件がユースケース spec に混在している箇所を洗い出す
- 共通要件を置く spec の分割方針を定義し、必要な共通 spec を新設する
- `AGENTS.md` の参照ルールを、整理後の spec 構成に合わせて更新する

## Capabilities

### New Capabilities
- `spec-structure`: architecture 文書と OpenSpec 文書群の責務境界、および共通要件の配置ルールを定義する

### Modified Capabilities

## Impact

- 影響コード: なし
- 影響文書: `openspec/specs/architecture.md`, `AGENTS.md`, 新設または再編される共通 spec
- 実装影響: 後続 change が参照する仕様書構成の前提になる
