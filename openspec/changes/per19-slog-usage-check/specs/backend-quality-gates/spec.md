## ADDED Requirements

### Requirement: slog利用規約違反を品質ゲートで検出しなければならない
システムは、`pkg/**` を対象に `slog` 利用規約違反を検出する静的チェックをバックエンド品質ゲートへ含めなければならない。検査は、`slog.Info/Error/Warn/Debug` の直接利用、`slog.*Context` を使うべき箇所での context なし logging、および主要構造化 key の lower_snake_case 違反を対象に含めなければならない。

#### Scenario: Contextなしのslog関数を直接利用する
- **WHEN** 開発者が `pkg/**` の業務処理で `slog.Info`、`slog.Error`、`slog.Warn`、`slog.Debug` を直接利用する
- **THEN** 品質ゲートは `slog.*Context` を使うべき違反として報告しなければならない

#### Scenario: 主要ログkeyがlower_snake_caseでない
- **WHEN** 開発者が `taskId` や `recordCount` のように lower_snake_case でない主要構造化 key を `slog` 呼び出しへ渡す
- **THEN** 品質ゲートは key 命名違反として報告しなければならない

#### Scenario: wrapper経由でcontext付きloggingする
- **WHEN** 開発者が `slog.*Context` を内部で利用する logger wrapper を経由して logging する
- **THEN** 品質ゲートは wrapper を禁止パターンとして誤報してはならない
