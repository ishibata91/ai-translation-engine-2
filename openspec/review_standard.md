# OpenSpec Review Standard

このファイルは、OpenSpec change を含む実装レビューで共通に使う観点だけを定義する。

## 1. 目的

- 変更差分ごとのレビュー観点のばらつきを減らす。
- 実装レビューに、仕様準拠だけでなく運用品質・保守性・AIデバッグ容易性の観点を含める。
- change 固有の review checklist を書く際の共通土台とする。

## 2. 基本原則

- `openspec/specs/architecture.md` の Interface-First AIDD / Vertical Slice 原則に一致すること。
- レビューは「動くか」だけでなく、「将来の修正・検証・障害解析が速くなるか」を見ること。
- 観点は共通ルールのみを扱い、change ごとの個別要件は別途 `review.md` に切り出すこと。
- change ごとの `review.md` は `openspec/review_template.md` を叩き台にして作ることを推奨する。

## 3. 共通観点

- 既存のパッケージ構成、命名、責務分離と整合しているか。
- 実装都合で architecture / design の方針から逸脱していないか。
- 変更差分に対して lint / format / test の標準品質ゲートを通せる状態か。
- change 固有観点は `openspec/changes/<change>/review.md` に切り出されているか。

## 4. バックエンド専用観点

`openspec/specs/backend_coding_standards.md` を基準に確認する。

### 4.1 アーキテクチャ・責務

- 契約・DTO・Mapper・Orchestrator の責務境界が崩れていないか。
- Slice 間の依存が Contract 越しになっているか。
- 業務ロジックや DTO の安易な横断共通化を持ち込んでいないか。
- `context.Context` が公開入口から末端まで流れているか。

### 4.2 ログ・テレメトリ

- AI デバッグ速度の向上を目的に、ログは自然文ではなく構造化された観測データとして扱う。
- ログ・トレースは原則として `pkg/infrastructure/telemetry` を経由して扱う。
- `context.Context` を入口から末端まで伝播し、`slog.*Context` を使ってテレメトリ情報を失わないこと。

- `pkg/infrastructure/telemetry` を使うべき箇所で独自 logging 実装を増やしていないか。
- `trace_id` / `span_id` や action 系メタデータが流れる設計になっているか。
- 外部 I/O、重要分岐、状態変化、異常系で、AI が追跡可能な構造化ログがあるか。
- 単純 helper や自明な処理に機械的なノイズログを増やしていないか。
- ログメッセージだけでなく、`slice`, `action`, `resource_type`, `resource_id`, `status`, `duration_ms`, `error_code` などのキーで機械的に解析できるか。

- `fmt.Println` 等の非構造化出力をデバッグ導線として持ち込まない。
- 全関数先頭に一律ログを入れるような、ノイズ優先の実装を推奨しない。
- telemetry 経由で追跡できない重要処理を新規追加しない。

### 4.3 実装品質

- 公開 API にのみ doc コメントがあるか。
- `error` が十分な文脈付きで wrap されているか。
- テストが主要シナリオと失敗系を十分にカバーしているか。

## 5. フロントエンド専用観点

`openspec/specs/frontend_architecture.md` と `openspec/specs/frontend_coding_standards.md` を基準に確認する。

### 5.1 Headless Architecture / 境界

- `pages` が描画責務に専念し、Wails API や複雑な状態管理を直接持っていないか。
- Wails API (`wailsjs/go/...`) を `hooks/features/*` からのみ呼んでいるか。
- `pages` から `wailsjs` や `store` を直接 import していないか。
- 機能専用型が `src/types` ではなく `src/hooks/features/<feature>/types.ts` 側へ局所化されているか。
- Hook の戻り値が UI に対する契約として整理され、state/action が読める形になっているか。

### 5.2 型・状態・副作用

- `any` を持ち込んでいないか。
- Wails や外部境界値を `unknown` + ランタイムパースで扱えているか。
- 非同期処理のフェーズが明示され、エラーを握りつぶしていないか。
- 副作用が `useEffect` に閉じ、イベント購読やタイマーにクリーンアップがあるか。
- `useEffect` 依存配列由来の多段発火へ依存していないか。

### 5.3 UI / 実装品質

- UI コンポーネントが表示責務に限定され、ビジネスロジックや Wails 呼び出しを直接持っていないか。
- Hook / コンポーネント / feature 型に必要な TSDoc があるか。
- Wails レスポンス整形が adapter に閉じ、UI で `as any` 吸収していないか。
- 巨大 JSX、巨大 hook、巨大関数を放置していないか。
- 変更中ファイルで `npm run lint:file -- <file...>` を使い、最後に `npm run lint:frontend` を通せる状態か。

## 6. change 固有観点との分離

- `openspec/changes/<change>/review.md` には、その change で特に確認すべき個別観点のみを書く。
- `openspec/changes/<change>/review.md` の構成は `openspec/review_template.md` を参考にする。
- tasks / spec / design の達成状況そのものは本ファイルの責務ではない。
- 検証手順、優先度分類、レポート形式は review standard ではなく verify 側の責務とする。
