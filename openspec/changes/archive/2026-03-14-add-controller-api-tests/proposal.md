## Why

`pkg/controller` には公開メソッドを持つ controller が複数存在する一方で、API テストの適用は一部に留まっており、既存挙動の退行を検知しにくい。既に定義済みの `api-test` 基盤要件を、現行 controller 群に対する具体的なテスト整備まで進める必要がある。

## What Changes

- 既存 `pkg/controller` 配下の公開メソッドを棚卸しし、API テスト未整備の controller を対象に table-driven test を追加する。
- controller ごとの依存差分を既存の API テスト補助方針に沿って整理し、共通化し過ぎずに再利用可能な test builder を整備する。
- 正常系と主要な異常系を中心に、既存 controller API の期待挙動を回帰テストとして固定する。
- バックエンド品質確認導線で既存 controller API テストが継続実行される前提を明確化する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `api-test`: `pkg/controller` に存在する現行 controller の公開メソッドへ API テストを段階的ではなく実運用に必要な基準で追加し、正常系と主要異常系を回帰検知できるようにする要件を追加する

## Impact

- 影響コードは `pkg/controller/**` の既存 controller、対応する API テスト、`pkg/tests/api_tests/**` の testenv / builder に及ぶ。
- 既存公開 API の外部仕様変更は想定せず、主な変更はテスト追加とテスト補助構成の整理である。
- 依存ライブラリは既存標準の `stretchr/testify` を継続利用し、新規に非デファクト標準ライブラリは導入しない。
