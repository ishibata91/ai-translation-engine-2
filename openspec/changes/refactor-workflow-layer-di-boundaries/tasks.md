> この task list は分割前の親一覧であり、実装は以下の 3 change の task を使う。
> - `reorganize-architecture-spec-boundaries`
> - `introduce-runtime-gateway-boundaries`
> - `migrate-master-persona-to-workflow`

## 1. 文書と責務区分の確定

- [ ] 1.1 `openspec/specs/architecture.md` の責務区分を実装方針へ反映し、`controller / workflow / usecase slice / runtime / gateway` の用語を他の関連 spec と整合させる
- [ ] 1.2 `openspec/specs/task/spec.md` `openspec/specs/persona/spec.md` `openspec/specs/queue/spec.md` の本文を `runtime / gateway` 前提へ更新する
- [ ] 1.3 必要であれば `AGENTS.md` と関連 OpenSpec 文書の参照先を調整し、アーキテクチャ文書と品質文書の責務重複を解消する

## 2. 契約とパッケージ境界の再定義

- [ ] 2.1 `pkg/controller` を新設または既存 Wails binding を移管し、controller が workflow 契約だけへ依存する入口を定義する
- [ ] 2.2 `pkg/workflow` を新設し、MasterPersona の開始・再開・キャンセル・状態管理を担う契約と state 管理を定義する
- [ ] 2.3 `pkg/persona` から runtime 制御責務を外し、workflow から呼び出される usecase slice 契約へ整理する
- [ ] 2.4 `pkg/runtime` と `pkg/gateway` の責務に沿って既存 `pkg/infrastructure/*` の移動方針を決め、公開 contract を整理する

## 3. MasterPersona 経路の再配置

- [ ] 3.1 MasterPersona の開始経路を `controller -> workflow -> persona / runtime` へ移し、controller が queue や persona 実装を直接触らないようにする
- [ ] 3.2 MasterPersona の resume / cancel / progress / phase 更新を workflow 主導へ寄せ、runtime が実行制御だけを担うようにする
- [ ] 3.3 parser 出力 -> persona 入力 DTO、および runtime 結果 -> persona 保存 DTO のマッピングを workflow へ集約する
- [ ] 3.4 既存 `task` API の互換性を保ちながら内部接続先を workflow へ差し替える

## 4. Runtime / Gateway の実装整理

- [ ] 4.1 `queue` `progress` `telemetry` `workflow state` のうち実行制御に属する要素を `runtime` として再配置する
- [ ] 4.2 `llm` `datastore` `config` `secret` など外部依頼に属する要素を `gateway` として再配置する
- [ ] 4.3 `runtime -> gateway` の限定依存として queue worker から LLM gateway を利用する contract を定義し、slice 固有ロジックが runtime に流入しないことを確認する
- [ ] 4.4 `main.go` または Wire injector の composition root を更新し、具象型配線を interface 経由へ寄せる

## 5. 品質ゲートと検証

- [ ] 5.1 `go-cleanarch` を導入し、責務区分ごとの依存方向ルールを定義する
- [ ] 5.2 `backend-quality-gates` の導線へ `go-cleanarch` を追加し、ローカル実行コマンドと CI 運用方針を更新する
- [ ] 5.3 変更対象ファイルごとに `npm run backend:lint:file -- <file...>` を回し、最終的に `npm run lint:backend` を通す
- [ ] 5.4 MasterPersona の開始・再開・キャンセル・完了 cleanup の経路をテストで確認し、runtime / gateway 分離後も既存の振る舞いが維持されることを検証する

## 6. 仕様書構成の再編

- [ ] 6.1 既存の OpenSpec 文書群を棚卸しし、UI 要件、workflow 要件、runtime 要件、gateway 要件がユースケース spec に混在している箇所を洗い出す
- [ ] 6.2 `architecture.md` の責務区分に合わせて spec の分割方針を定義し、ユースケース単位 spec から UI / runtime / gateway の共通要件を分離する
- [ ] 6.3 必要な共通 spec を新設または既存 spec を再編し、今後はユースケース spec がユースケース固有要件に集中できる構成へ整理する
