## 1. Controller 境界の整理

- [x] 1.1 既存違反を固定リストとして扱い、`main.App` の `DictGetSources`、`DictDeleteSource`、`DictGetEntries`、`DictGetEntriesPaginated`、`DictSearchAllEntriesPaginated`、`DictUpdateEntry`、`DictDeleteEntry`、`DictStartImport`、`SelectFiles`、`SelectJSONFile`、`Greet`、`SetConfigService`、`SetDictService` を controller 移管または削除対象として整理する
- [x] 1.2 `main.go` の `Bind` に含まれる controller 外 bind 3 件、`main.App`、`pkg/modelcatalog.ModelCatalogService`、`pkg/persona.Service` を解消対象として固定し、変更後は controller のみになるよう設計を反映する
- [x] 1.3 辞書 API とファイルダイアログ API を受ける controller を `pkg/controller` に追加または既存 controller へ整理し、入力整形と service 委譲だけを持つ構成にする
- [x] 1.4 `modelcatalog` の `ListModels` と `persona` の `ListNPCs`、`ListDialoguesByPersonaID` を受ける controller を `pkg/controller` に追加し、slice service の直接 bind を不要にする

## 2. Wails バインディングの差し替え

- [x] 2.1 `main.go` の DI を更新し、新しい controller 群を生成して `OnStartup` で必要な `context.Context` を注入する
- [x] 2.2 `main.go` の `Bind` から `main.App`、`pkg/modelcatalog.ModelCatalogService`、`pkg/persona.Service` を外し、`pkg/controller` 配下の型のみを公開する
- [x] 2.3 `app.go` をライフサイクル保持専用へ縮退させ、辞書ラッパーや旧 bind 用メソッドを削除する
- [x] 2.4 初期テンプレート由来の `App.Greet` を今回の change で削除するかを決め、削除する場合は関連する Wails 生成物と参照を合わせて整理する

## 3. Frontend の追従

- [x] 3.1 `frontend/src/wailsjs/go/main/App.*`、`frontend/src/wailsjs/go/modelcatalog/ModelCatalogService.*`、`frontend/src/wailsjs/go/persona/Service.*` の参照箇所を洗い出す
- [x] 3.2 Wails 生成物を再生成し、feature hook / adapter の import を新しい controller ベースの `wailsjs` 参照へ一括で置き換える
- [x] 3.3 `pages` からの直接 `wailsjs` import が増えていないことを確認し、必要な修正を `hooks/features/*` 側に閉じ込める

## 4. 検証と品質ゲート

- [x] 4.1 変更したバックエンドファイルごとに `npm run backend:lint:file -- <file...>` を実行し、違反を解消しながら進める
- [x] 4.2 変更したフロントエンドファイルごとに `npm run lint:file -- <file...>` を実行し、`wailsjs` 参照変更による lint 違反を解消する
- [x] 4.3 `npm run lint:backend` を実行し、Wails bind の再配置後もバックエンド品質ゲートを満たすことを確認する
- [x] 4.4 `npm run lint:frontend` を実行し、frontend が新 controller API 参照で成立していることを確認する
