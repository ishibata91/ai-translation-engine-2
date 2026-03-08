## Why

フロントエンドのリファクタ方針は定義されている一方で、TSDoc の記述漏れや TypeScript の公開範囲肥大化を機械的に検出する品質ゲートが不足している。今後の継続的なリファクタで規約を運用可能にするため、編集中ファイルを中心に素早く回せる lint 導線とあわせて仕様化する必要がある。

## What Changes

- フロントエンドのコード品質を機械検証する capability を追加し、TSDoc 必須箇所の検査を lint で強制できるようにする。
- TypeScript でファイル外に公開する識別子を必要最小限に抑えるルールを追加し、不要な `export` を検出または抑制できるようにする。
- 全体 lint に加えて、通常のリファクタ作業では編集中ファイルだけを対象に素早く実行できる lint フローを定義する。
- デファクトスタンダードの lint エコシステムを前提とし、既存 ESLint 基盤に統合できる運用を明確化する。

## Capabilities

### New Capabilities
- `frontend-code-quality-guardrails`: フロントエンドの doc コメント検査、公開範囲最小化、ファイル単位 lint 実行を含む品質ゲートを定義する

### Modified Capabilities
- `frontend-headless-architecture`: Headless Architecture を維持するために必要な lint 境界検証と運用導線の要求を補強する

## Impact

- 対象コード: `frontend/src` 配下の TypeScript / React コード、lint 設定、npm scripts
- 影響システム: フロントエンド開発フロー、CI/ローカル検証フロー
- 依存候補: `eslint`, `typescript-eslint`, `eslint-plugin-import`, `eslint-plugin-jsdoc` または TSDoc 系のデファクトツール群
