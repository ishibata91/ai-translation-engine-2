## 1. workflow 契約導入

- [x] 1.1 `pkg/workflow` の公開 contract を定義する
- [x] 1.2 controller が依存する workflow 入口 contract を定義する

## 2. runtime / gateway 境界導入

- [x] 2.1 `queue` `progress` `workflow state` を runtime として再整理する
- [x] 2.2 `llm` `datastore` `config` などを gateway として再整理する
- [x] 2.3 `runtime -> gateway` の限定依存として queue worker -> llm gateway の contract を定義する

## 3. 品質ゲート

- [x] 3.1 `go-cleanarch` を導入する
- [x] 3.2 `backend-quality-gates` の導線へ `go-cleanarch` を追加する
- [x] 3.3 境界導入後に `backend:lint:file` と `lint:backend` で検証する
