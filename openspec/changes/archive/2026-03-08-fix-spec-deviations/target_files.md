# 修正対象ファイルリスト

このチェンジ（`fix-spec-deviations`）で修正対象となる主要なファイルとその違反内容をリストアップします。
（※各ファイルの違反箇所・行数は2026-03-08時点のものです）

## 1. バックエンド: Context 伝播違反（Backend Coding Standards Compliance）

### `pkg/workflow/master_persona_service.go`
- **対象行**: 192行目付近 `StartMasterPersonTask`
- **対象行**: 199行目付近 `ResumeMasterPersonaTask`
- **対象行**: 204行目付近 `CancelTask`
- **違反内容**: 公開メソッドの第一引数で `ctx context.Context` を受け取っておらず、内部で `context.Background()` をハードコードして生成・使用しているため、TraceID等の文脈伝播が途切れてしまう（アーキテクチャ規約違反）。

## 2. バックエンド: LM Studioの Structured Output 解析処理

### `pkg/infrastructure/llm/local_client.go`
- **対象行**: 323行目付近 `json.Unmarshal(raw.Choices[0].Message.Content, &content)`
- **違反内容**: LM Studio に対して `response_format: {"type": "json_schema"}` を指定してリクエストした際、レスポンスの `content` がエスケープされた文字列ではなく JSON オブジェクトそのものとして返却されるケースがある。現在の実装では `string` へのデコードを前提としているため、パースエラー（Unmarshal failed）が発生し、アプリが進行不能になる可能性が高い。

## 3. フロントエンド: UI層の VSA / 責務分離違反 (Frontend Coding Standards Compliance)

### `frontend/src/components/ModelSettings.tsx`
- **対象行**: 4行目 `import { ListModels } from '../wailsjs/go/modelcatalog/ModelCatalogService';`
- **対象行**: 440行目付近 `ListModels(...)` の直接呼び出し
- **違反内容**: UIコンポーネントディレクトリ配下の表示コンポーネントが、バックエンド通信エンドポイント（`wailsjs`）を直接import・実行している。Headless Architecture（VSA）の原則に従い、通信ロジックは `hooks/features/` 側に寄せる必要がある。

### `frontend/src/components/log/LogViewer.tsx`
- **対象行**: 4行目 `import { UIStateGetJSON, UIStateSetJSON } from '../../wailsjs/go/config/ConfigService';`
- **対象行**: 774行目・798行目付近での直接呼び出し
- **違反内容**: 同じく UI コンポーネントが `wailsjs` を直接呼び出しており、責務（表示と通信ロジック）の分離ができていない。設定永続化処理などはカスタムHook側に移譲すべき。
