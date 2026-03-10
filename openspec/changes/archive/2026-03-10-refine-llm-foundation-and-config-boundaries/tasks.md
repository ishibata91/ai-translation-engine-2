## 1. foundation 境界の定義

- [x] 1.1 `architecture.md` に `foundation` 区分を追加し、`telemetry` と `progress` を横断基盤として扱う依存方向を明文化する
- [x] 1.2 `telemetry` と `progress` の責務を整理し、transport / observability と workflow 側の意味づけを分離して記述する
- [x] 1.3 この change の対象が LLM 周辺に限定されることを review / design に明記する

## 2. 品質ゲート更新

- [x] 2.1 `backend-quality-gates` spec に foundation 区分を反映し、`depguard` の期待動作を更新する
- [x] 2.2 `.golangci.yml` の files ルールへ foundation を追加し、各区分から foundation への依存と foundation からの逆依存を整理する
- [x] 2.3 LLM 周辺で検出される runtime 依存違反が foundation 移行で解消できることを確認する

## 3. LLM 周辺の移行

- [x] 3.1 `telemetry` を foundation 配下へ移し、`pkg/gateway/llm/**` の runtime 依存を解消する
- [x] 3.2 `progress` を foundation 配下へ移し、LLM 周辺の notifier 依存を runtime から切り離す
- [x] 3.3 LLM 周辺の `controller` / `workflow` / `runtime` / `gateway` の import を更新し、event 名や logger 挙動が維持されることを確認する

## 4. 検証

- [x] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行する
- [x] 4.2 `npm run lint:backend` を実行し、foundation 追加後の依存方向違反が想定どおり収束することを確認する
- [x] 4.3 LLM 周辺の Go テストを実行し、Wails logger / progress notifier を含む回帰がないことを確認する
- [x] 4.4 repository 全体への展開は後続 change とし、この change が LLM 周辺限定であることを review に残す
