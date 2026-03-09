## Why

現状の Wails バインディングは `main` や usecase slice 側に分散しており、`architecture.md` が定義する「Wails binding は `pkg/controller` の責務」という境界と実装が一致していない。`PER-21` で指摘された公開面のばらつきを解消し、UI 向け API の露出範囲を controller 経由に限定するため、今このタイミングで整理が必要である。

## What Changes

- `main.go` の Wails `Bind` を見直し、`pkg/controller` 配下の controller のみを公開対象にする。
- `main.App` が持つ辞書系ラッパーや UI 向け公開メソッドを controller へ移管し、`main` から UI-facing API を外す。
- `pkg/modelcatalog.ModelCatalogService` と `pkg/persona.Service` の Wails 直接公開をやめ、controller 経由の公開 API に統一する。
- `frontend/src/wailsjs` の生成物変更を前提に、フロントエンドの呼び出し先を controller ベースの API 参照へ追従させる。
- Wails 公開 API の責務境界を OpenSpec に反映し、今後 `pkg/controller` 外を直接 bind しない前提を明文化する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `wails-app-shell`: Wails が公開する Go API は `pkg/controller` 配下の controller に限定し、`main` や usecase slice の service を直接 bind しない requirement へ更新する

## Impact

- 影響コード: `main.go`, `app.go`, `pkg/controller/*`, `pkg/modelcatalog/service.go`, `pkg/persona/persona_service.go`
- 影響 API: Wails の公開オブジェクト名、`frontend/src/wailsjs/go/*` の生成先、フロントエンドの import / 呼び出し先
- 影響システム: Wails 起動時 DI 構成、React フロントエンドの Go バインディング利用箇所
- 依存追加は想定しない。既存の Wails / React / Go 構成のまま責務再配置で対応する
