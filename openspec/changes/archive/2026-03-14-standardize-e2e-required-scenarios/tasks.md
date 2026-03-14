## 1. OpenSpec とテスト資産の整理

- [x] 1.1 `openspec/specs/e2e-required-scenarios/spec.md` を追加し、ページ単位 E2E の標準ルールを main specs に反映する
- [x] 1.2 `openspec/specs/e2e-required-scenarios/dictionary-builder/spec.md` を追加し、`DictionaryBuilder` の必須シナリオを main specs に反映する
- [x] 1.3 `openspec/specs/playwright-quality-gate/spec.md` を更新し、ページ単位 E2E の合格条件を `必須シナリオ` 基準へ揃える
- [x] 1.4 `frontend/src/e2e/fixtures/dictionary-builder/` を作成し、`DictionaryBuilder` 用 mock XML と補助 fixture の置き場を揃える

## 2. E2E モックと PageObject の拡張

- [x] 2.1 `frontend/src/e2e/helpers/wails-mock.ts` を拡張し、辞書ソース一覧、エントリ一覧、横断検索結果の固定データを返せるようにする
- [x] 2.2 mock XML を `frontend/src/e2e/fixtures/dictionary-builder/` に配置し、E2E から参照しやすい命名に揃える
- [x] 2.3 `frontend/src/e2e/page-objects/pages/dictionary-builder.po.ts` に一覧確認、ソース選択、詳細確認、編集画面遷移、横断検索操作の API を追加する
- [x] 2.4 必要なら `frontend/src/e2e/fixtures/app.fixture.ts` を更新し、`DictionaryBuilder` シナリオに必要な fixture 初期化を整理する

## 3. DictionaryBuilder 必須シナリオ E2E 実装

- [x] 3.1 `DictionaryBuilder` 向け E2E spec を追加し、一覧表示の必須シナリオを実装する
- [x] 3.2 `DictionaryBuilder` 向け E2E spec に、ソース選択から詳細/編集導線の必須シナリオを実装する
- [x] 3.3 `DictionaryBuilder` 向け E2E spec に、横断検索導線の必須シナリオを実装する
- [x] 3.4 spec 側で locator を直接使わず、PageObject API 経由だけで完結していることを確認する

## 4. 品質ゲート確認

- [x] 4.1 変更対象の E2E 関連ファイルに対して `npm run lint:file -- <file...>` を実行し、指摘を解消する
- [x] 4.2 `npm run lint:frontend` を実行し、frontend 全体の lint を通す
- [x] 4.3 `npm run e2e` を実行し、`DictionaryBuilder` の必須シナリオが品質ゲートとして通ることを確認する
- [x] 4.4 `openspec/changes/standardize-e2e-required-scenarios/review.md` に完了条件と実行結果を記録する
