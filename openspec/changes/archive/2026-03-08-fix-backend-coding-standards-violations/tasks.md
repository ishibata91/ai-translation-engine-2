## 1. task / workflow の規約違反是正

- [x] 1.1 `pkg/task/bridge.go` で `context.Background()` を使っている呼び出しを、保持済み ctx または引数 ctx を使う形へ修正する
- [x] 1.2 `pkg/task/manager.go` で store / hook / goroutine 起点に流す ctx を整理し、`context.Background()` 乱用を解消する
- [x] 1.3 `pkg/task/manager.go` の error return と logging を見直し、wrap と `slog.*Context` を統一する
- [x] 1.4 `pkg/workflow/master_persona_service.go` で受け取った ctx を破棄している箇所を修正し、resume / cancel / queue / save 経路へ伝播させる
- [x] 1.5 `pkg/workflow/master_persona_service.go` の error wrap 欠落と握りつぶしを是正し、必要な warning/error ログを追加する
- [x] 1.6 `pkg/task/manager.go` と `pkg/workflow/master_persona_service.go` の責務過多メソッドを private method 抽出で分割する
- [x] 1.7 `npm run backend:lint:file -- pkg/task/bridge.go pkg/task/manager.go pkg/workflow/master_persona_service.go` を実行し、違反が残らないことを確認する

## 2. config の binding 入口整理

- [x] 2.1 `pkg/config/config_service.go` を `ConfigController` 方針で整理し、命名と責務が controller 入口として読める状態にする
- [x] 2.2 `pkg/config` の Wails binding 入口から store 呼び出しまで ctx を伝播させ、`context.Background()` 固定を解消する
- [x] 2.3 `pkg/config` の公開メソッドで返す error に namespace / key 文脈を付与する
- [x] 2.4 `pkg/config` の logging が必要な箇所を `slog.*Context` と fixed key にそろえる
- [x] 2.5 `npm run backend:lint:file -- pkg/config/config_service.go pkg/config/sqlite_store.go` を実行し、違反が残らないことを確認する

## 3. pipeline の error / log / cleanup 是正

- [x] 3.1 `pkg/pipeline/store.go` の DB エラー返却に文脈付き wrap を追加する
- [x] 3.2 `pkg/pipeline/manager.go` の logging を `slog.*Context` と識別子ベースのメッセージへ統一する
- [x] 3.3 `pkg/pipeline/manager.go` と `pkg/pipeline/handler.go` の握りつぶしと cleanup failure を整理し、継続時も warning を残す
- [x] 3.4 `pkg/pipeline/manager.go` の公開メソッド内の責務過多部分を private method 抽出で整理する
- [x] 3.5 `npm run backend:lint:file -- pkg/pipeline/store.go pkg/pipeline/manager.go pkg/pipeline/handler.go` を実行し、違反が残らないことを確認する

## 4. テストと再発防止

- [x] 4.1 `context` 伝播、error wrap、resume / cancel、cleanup failure を対象に、関連 package の Go テストを Table-Driven で追加または更新する
- [x] 4.2 `violation-targets.md` の高優先度項目がコード上で解消されたことを確認し、必要ならメモを更新する
- [x] 4.3 `tools/backendquality` に低コストで追加できる静的検査候補を確認し、今回スコープで入れるか見送るかを判断する
- [x] 4.4 `npm run lint:backend` を実行し、バックエンド全体の lint が通ることを確認する
- [x] 4.5 関連する Go テストを実行し、変更範囲で回帰がないことを確認する
