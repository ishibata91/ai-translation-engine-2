## Why

現在の Playwright 品質ゲートはアプリシェルと主要画面遷移の最低限確認に留まり、ページ単位で何を通れば品質ゲート合格とみなすかが共通仕様として定義されていない。辞書構築画面の E2E を追加する前に、画面固有文脈に依存しない標準的な「必須シナリオ」の定義を先に整備し、今後のページ追加でも同じ基準で品質ゲートを拡張できる状態にする必要がある。

## What Changes

- ページ単位 E2E の合格条件を、画面固有名称ではなく共通的な `必須シナリオ` の概念で定義する新しい標準 spec を追加する。
- Playwright 品質ゲート spec を更新し、ページ単位の回帰検知を `必須シナリオ` ベースで記述できるようにする。
- `DictionaryBuilder` を最初の適用対象として、一覧表示、選択後の詳細/編集導線、横断検索結果導線をページ単位の必須シナリオとして E2E に追加する。
- 追加する E2E は既存方針どおり PageObject 中心で実装し、spec から locator を直接拡散させない。

## Capabilities

### New Capabilities
- `e2e-required-scenarios`: ページ単位 E2E の必須シナリオ、合格条件、シナリオ粒度を標準化する。親 spec を標準定義とし、ページ別要件は配下ディレクトリへ分ける。

### Modified Capabilities
- `playwright-quality-gate`: 品質ゲート要件を画面固有の表現から `必須シナリオ` ベースへ一般化し、ページ単位 E2E の合格条件を明示する。

## Impact

- OpenSpec: `openspec/specs/playwright-quality-gate/spec.md` の要件更新、および新規 `openspec/specs/e2e-required-scenarios/spec.md` とページ別配下 spec の追加。
- Frontend E2E: `frontend/src/e2e` 配下の fixture、PageObject、`DictionaryBuilder` 向け spec の追加または拡張。
- 品質ゲート運用: `npm run e2e` により、`DictionaryBuilder` のページ単位 E2E を回帰検知の必須ケースとして扱う。