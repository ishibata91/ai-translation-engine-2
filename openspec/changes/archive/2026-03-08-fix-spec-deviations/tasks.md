<artifact id="tasks" change="fix-spec-deviations" schema="spec-driven">

## 1. LM Studio Structured Output Parsing (`pkg/infrastructure/llm/local_client.go`)
- [x] 1.1 `local_client.go` における `doChatCompletion` メソッドの実装を修正する。
  - **何がだめか**: LM Studio のローカル LLM から Structured Output をリクエストした際、レスポンスの `content` がエスケープされた文字列ではなくパース済みの JSON オブジェクトとして返ってくるケースなどがあり、現在の `json.Unmarshal(raw.Choices[0].Message.Content, &content)` （文字列へのアンマーシャル）でデコードエラーが発生してパースに失敗している可能性がある。または `json_schema` のペイロード形式が LM Studio と非互換でエラーになっている。
  - **どうする**: レスポンスの `content` が文字列であるかオブジェクトであるかを型のフォールバック等を用いて堅牢にパースできるように修正する。またリクエスト時の `response_format` が LM Studio 互換の正しい payload になるよう調整し、正しく JSON 文字列（またはオブジェクト）を取得・返却できるようにする。

## 2. Backend Coding Standards Compliance (`pkg/` 以下の Context / slog)
- [x] 2.1 `context.Context` の伝播漏れの修正
  - **何がだめか**: `pkg/` 以下のいくつかの関数やメソッド（I/O を伴うものや外部APIを叩くもの）において、第一引数に `ctx context.Context` を受け取っていない、または上位から下位へ `ctx` がリレーされていない箇所がある（コンテキスト設計の標準に違反）。
  - **どうする**: 全パッケージを点検し、必要な関数シグネチャに `ctx context.Context` を追加する。呼び出し元もすべて `ctx` を渡すように修正する。
- [x] 2.2 構造化ログ (`slog.*Context`) の徹底
  - **何がだめか**: コードの随所で `slog.Info()` や `slog.Error()` が使用されており、コンテキスト情報（TraceIDなど）を引き回すための `slog.InfoContext(ctx, ...)` が使われていない。
  - **どうする**: `pkg/` 以下のすべてのログ呼び出しを検索し、`ctx` を第一引数に取る `InfoContext`, `ErrorContext`, `DebugContext`, `WarnContext` に置き換える。
- [x] 2.3 `npm run backend:lint` を実行・通過させる。

## 3. Frontend Coding Standards Compliance (Headless Components / VSA)
- [x] 3.1 Components 内の `wailsjs` 依存の排除
  - **何がだめか**: VSA および Headless Architecture パターンにおいて、UIコンポーネント（表示責務）はバックエンド通信（インフラロジック）を直接持つべきではないが、`frontend/src/components/` 配下（例: `ModelSettings.tsx`, `LogViewer.tsx` など）で `wailsjs` が直接インポートされ利用されている。
  - **どうする**: `wailsjs` を叩くロジックを `hooks/features/` 配下の Custom Hook（例: `useModelSettings.ts`, `useLogViewer.ts`）に切り出す。コンポーネント側はそれらの Hook を呼び出して State や関数を受け取るだけの「純粋な表示層」へとリファクタリングする。
- [x] 3.2 境界値の `any` 削除と `valibot` / `zod` バリデーションの徹底
  - **何がだめか**: バックエンドとの境界（Wailsのレスポンスやイベントリスナ）で `any` で受け取っている箇所が一部に残存している可能性がある。
  - **どうする**: 境界値は `unknown` で受け取り、スキーマバリデーションライブラリで安全にパースしてからアプリケーション内に取り込むよう修正する。
- [x] 3.3 `npm run lint:frontend` を実行し、全アラートを解消する。

## 4. Final Verification
- [x] 4.1 全モジュールのテストを通す（`npm run backend:check`, `npm run build`）。
- [x] 4.2 本変更でデグレが起きていないか、APIとUIの結合動作を確認する。

</artifact>
