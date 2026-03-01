## 1. Backend (Telemetry Naming & Context)

- [x] 1.1 `pkg/infrastructure/telemetry/context.go` の `WithRequestID` 関数を `WithTraceID` にリネームする
- [x] 1.2 `pkg/infrastructure/telemetry` の文字列 `"request_id"` をすべて `"trace_id"` に置換する
- [x] 1.3 `app.go` などのWailsバインディング呼び出しでの `WithRequestID` 使用箇所をすべて `WithTraceID` に変更する

## 2. Backend (Log Streaming)

- [x] 2.1 Wailsのイベントシステム (`runtime.EventsEmit`) にログを転送するための新しい `slog.Handler` (例: `wailsHandler`) を実装する
- [x] 2.2 `wailsHandler` に必要な Wails ランタイム `context.Context` を初期化プラミングする（`app.go` の `startup` 時などで注入する）
- [x] 2.3 `telemetry.ProvideLogger` の構成を更新し、コンソール出力だけでなく新規作成した `wailsHandler` へもログをブロードキャストする

## 3. Backend (Config Service Binding)

- [x] 3.1 フロントエンドからUI状態の永続化を行えるよう、`pkg/config.UIStateStore` の Wails ラッパー (`ConfigService` 等) を実装する
- [x] 3.2 実装したラッパーを `main.go` の `wails.Run` メソッドの `Bind` プロパティに追加する

## 4. Frontend (Types & Log Data)

- [x] 4.1 フロントエンド用のログエントリ型（`LogEntry` / `TelemetryLog` 等）を定義し、バックエンドの構造化ログ（`trace_id`, `action`, `resource_type` などのフィールド）と同期させる

## 5. Frontend (Log Viewer Component & Streaming)

- [x] 5.1 `LogViewer.tsx` コンポーネントに `runtime.EventsOn` を用いて非同期ログ受信処理を追加する
- [x] 5.2 受信したログをReactステート内で配列として管理し、表示更新パフォーマンスのため必要に応じたバッファリングや間引きを実装する
- [x] 5.3 ログレベルおよび TraceID 文字列検索による一覧の絞り込み表示処理を実装する

## 6. Frontend (Log Detail Pane)

- [x] 6.1 リストからログエントリをクリックした際に動作する下部からのスライドアップパネル（Detail Pane）を実装する
- [x] 6.2 パネル内に、スタックトレースや `trace_id`, `action`, `resource_type` などの全てのテレメトリ属性を出力するようにレンダリングを更新する

## 7. Frontend (Filter Persistence)

- [x] 7.1 `LogViewer.tsx` にて、コンポーネントマウント時に Wails の Config サービスを通じて保存されたログフィルター設定（`UIStateStore`）を取得する処理を実装する
- [x] 7.2 保存された設定がない場合、ログレベル初期値を「ERROR」として設定・初期化する
- [x] 7.3 ユーザーがフィルター設定を変更した際、即座に Config サービスの `SetJSON` 等のメソッドを呼んで永続化する処理を実装する
