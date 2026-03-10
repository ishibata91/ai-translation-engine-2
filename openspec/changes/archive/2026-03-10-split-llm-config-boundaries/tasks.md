## 1. config 責務の棚卸し

- [x] 1.1 `workflow/config` の中身を store contract、SQLite 実装、migration、TypedAccessor、workflow 固有 default に分類する
- [x] 1.2 LLM 周辺で `workflow/config` と `gateway/config` を参照している箇所を洗い出す
- [x] 1.3 この change の対象が LLM 周辺に限定されることを review / design に明記する

## 2. package 分割設計

- [x] 2.1 `pkg/gateway/configstore` に置く contract / SQLite 実装 / migration の境界を確定する
- [x] 2.2 `pkg/runtime/configaccess` に置く TypedAccessor / 実行時読取補助の境界を確定する
- [x] 2.3 `pkg/workflow/persona` に置く persona slice 向け default / 解釈の境界を確定する

## 3. LLM 周辺の移行

- [x] 3.1 `pkg/gateway/config` の workflow 再エクスポートを廃止し、`gateway/configstore` 実装へ置き換える
- [x] 3.2 LLM 周辺の `runtime` / `workflow` / `controller` の import を `configstore` / `configaccess` / `workflow/persona` へ切り替える
- [x] 3.3 prompt default、secret access、model catalog、queue worker の設定利用が新しい境界で成立することを確認する

## 4. 検証

- [x] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行する
- [x] 4.2 `npm run lint:backend` を実行し、gateway -> workflow を含む config 境界違反が解消されたことを確認する
- [x] 4.3 LLM 周辺の Go テストを実行し、設定読取と prompt default の回帰がないことを確認する
- [x] 4.4 他ユースケースの config 移行は後続 change で扱う前提を review に残す
