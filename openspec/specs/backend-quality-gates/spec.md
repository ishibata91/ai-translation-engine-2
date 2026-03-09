# Purpose

バックエンド品質ゲートの必須ツール、lint 標準、実行導線、任意検査の位置付けを OpenSpec capability として定義する。

## Requirements

### Requirement: 品質ゲートの必須ツール定義
システムは、バックエンド品質ゲートの必須ツールとして `golangci-lint`、`goimports`、`goleak` を定義しなければならない。

#### Scenario: 必須ツールセットが明示される
- **WHEN** 品質ゲート定義を確認する
- **THEN** `golangci-lint`、`goimports`、`goleak` が必須実行項目として列挙されている

### Requirement: lintルールセットの標準化
システムは、`golangci-lint` で実行する標準lintセットを定義しなければならない。標準lintセットは少なくとも `staticcheck`、`govet`、`errcheck`、`revive`、`gosec`、`stylecheck` を含まなければならない。

#### Scenario: lint実行結果の一貫性が確保される
- **WHEN** 開発者が同一設定で lint を実行する
- **THEN** 実行対象lintとfail条件が統一され、環境差分による判定揺れが発生しない

#### Scenario: 公開シンボルのdoc欠落が検知される
- **WHEN** 開発者が公開型、公開関数、公開メソッドの doc コメントを欠いた状態で lint を実行する
- **THEN** `stylecheck` の `ST1020` / `ST1021` / `ST1022` により違反が報告される

### Requirement: 実行導線の定義
システムは、ローカルとCIの双方で同一品質ゲートを実行できる導線を提供しなければならない。

#### Scenario: ローカルとCIで同じ判定基準になる
- **WHEN** 開発者がローカル実行したチェックとCI実行結果を比較する
- **THEN** 同一コマンド群または同等設定で判定結果が一致する

### Requirement: 依存方向 lint を品質ゲートへ含めなければならない
システムは `go-cleanarch` を用いて責務区分の依存方向違反を検出し、バックエンド品質ゲートへ含めなければならない。

#### Scenario: 依存方向違反が検出される
- **WHEN** controller が runtime 具象へ直接依存するなどの違反を追加する
- **THEN** `go-cleanarch` は違反を検出しなければならない

#### Scenario: runtime から gateway の限定依存だけが許可される
- **WHEN** queue worker が LLM gateway を利用する
- **THEN** 品質ゲートは当該依存を許可しなければならない
- **AND** runtime から slice 具象への依存は許可してはならない

### Requirement: ファイル単位lint導線の提供
システムは、バックエンド変更中の反復修正を支えるため、指定した `.go` ファイル群だけを対象にした lint 実行導線を提供しなければならない。

#### Scenario: 変更ファイルだけを先に検査できる
- **WHEN** 開発者が `backend:lint:file` に変更ファイルを渡して実行する
- **THEN** 指定ファイルを含むパッケージを解決した上で、指定ファイルに紐づく lint 違反のみが報告される

### Requirement: lint設定の専用テストを要求しない
システムは、lint設定そのものに対する専用テストコードの作成を必須要件にしてはならない。lintは設定ファイルと実行コマンドにより運用しなければならない。

#### Scenario: lint運用がシンプルに維持される
- **WHEN** 開発者が品質ゲート要件を確認する
- **THEN** lint設定の単体テストや専用テストケース作成は必須化されていない

### Requirement: 脆弱性検査の任意運用
システムは、`govulncheck` を常時必須ゲートに含めず、依存更新時またはリリース前の任意検査として定義しなければならない。

#### Scenario: ローカル専用運用に合わせた検査強度を維持する
- **WHEN** 品質ゲートを日常開発で実行する
- **THEN** `govulncheck` は必須失敗条件に含まれず、任意実行手順として別途定義されている

### Requirement: context伝播違反を品質ゲートで検出しなければならない
システムは、`pkg/**` を対象に `context.Context` の伝播違反を検出する静的検査をバックエンド品質ゲートへ含めなければならない。検査は、公開入口で受けた `ctx` の未伝播、不適切な `context.Background()` / `context.TODO()` の利用、および goroutine 起点での `ctx` 脱落を検出できなければならない。

#### Scenario: 公開入口で受けたctxを内部処理へ渡さない
- **WHEN** 開発者が `ctx context.Context` を受けた公開メソッドから、I/O または外部依頼を行う内部処理へ `ctx` を渡さずに呼び出す
- **THEN** 品質ゲートは `context` 伝播違反として報告しなければならない

#### Scenario: context.BackgroundまたはTODOを業務処理で新規生成する
- **WHEN** 開発者が公開入口や workflow / slice / gateway の業務処理内で、既存の `ctx` を無視して `context.Background()` または `context.TODO()` を利用する
- **THEN** 品質ゲートは trace / cancel 伝播を壊す違反として報告しなければならない

#### Scenario: goroutine起点でctxを落とす
- **WHEN** 開発者が goroutine を起動する処理で、親 `ctx` を引き継がずに非同期処理を開始する
- **THEN** 品質ゲートは goroutine 境界での `context` 脱落として報告しなければならない

### Requirement: context伝播検査は正当な除外境界を持たなければならない
システムは、初期化専用コード、テスト補助、純粋関数など `context.Context` 伝播義務の外にあるケースを識別し、実運用不能な誤検知を避けなければならない。

#### Scenario: 初期化専用コードは一律違反にしない
- **WHEN** 開発者が composition root やテストセットアップで `context.Background()` を利用する
- **THEN** 品質ゲートは当該コードを除外対象として扱うか、少なくとも業務処理違反とは区別して報告しなければならない

### Requirement: slog利用規約違反を品質ゲートで検出しなければならない
システムは、`pkg/**` を対象に `slog` 利用規約違反を検出する静的チェックをバックエンド品質ゲートへ含めなければならない。検査は、`slog.Info/Error/Warn/Debug` の直接利用、`slog.*Context` を使うべき箇所での context なし logging、および主要構造化 key の lower_snake_case 違反を対象に含めなければならない。

#### Scenario: slog.Logger メソッドでも context なし利用を検出する
- **WHEN** 開発者が `*slog.Logger` に対して `Info/Warn/Error/Debug` を context なしで直接呼び出す
- **THEN** 品質ゲートは `*Context` 系メソッドを使うべき違反として報告しなければならない

#### Scenario: Contextなしのslog関数を直接利用する
- **WHEN** 開発者が `pkg/**` の業務処理で `slog.Info`、`slog.Error`、`slog.Warn`、`slog.Debug` を直接利用する
- **THEN** 品質ゲートは `slog.*Context` を使うべき違反として報告しなければならない

#### Scenario: 主要ログkeyがlower_snake_caseでない
- **WHEN** 開発者が `taskId` や `recordCount` のように lower_snake_case でない主要構造化 key を `slog` 呼び出しへ渡す
- **THEN** 品質ゲートは key 命名違反として報告しなければならない

#### Scenario: wrapper経由でcontext付きloggingする
- **WHEN** 開発者が `slog.*Context` を内部で利用する logger wrapper を経由して logging する
- **THEN** 品質ゲートは wrapper を禁止パターンとして誤報してはならない

#### Scenario: key 命名検査は文字列リテラルを対象にする
- **WHEN** 開発者が `slog.String(\"taskId\", value)` や `slog.InfoContext(ctx, msg, \"recordCount\", count)` のように文字列リテラル key を与える
- **THEN** 品質ゲートは lower_snake_case 違反を報告しなければならない
- **AND** 定数展開や動的生成 key は MVP の自動検査対象外として扱ってよい

### Requirement: error wrap不足を品質ゲートで検出しなければならない
システムは、package 境界または外部 I/O 境界で返却される `error` に文脈付き wrap が不足しているケースを、バックエンド品質ゲートで検出しなければならない。検査は、単純な `return err`、`fmt.Errorf` における `%w` 欠落、および cleanup 以外での失敗握りつぶしを対象に含めなければならない。

- package 境界には、公開関数、公開メソッド、および他 package から利用される公開 Contract 実装の返却点を含めなければならない。
- 外部 I/O 境界には、gateway / datastore / 外部 API / ファイル I/O の失敗を上位へ返す箇所を含めなければならない。
- `fmt.Errorf` 検査は、元 `error` を引数に含みつつ `%w` が使われていないケースに限定しなければならない。
- 品質ゲートは、同じ `error` 変数に対する `err = fmt.Errorf("...: %w", err)` の自己 wrap 再代入を検出しなければならない。

#### Scenario: package境界でreturn errを返す
- **WHEN** 開発者が package 境界の公開メソッドまたは外部依頼の呼び出し箇所で、取得した `err` を文脈付与せず `return err` する
- **THEN** 品質ゲートは error wrap 不足として報告しなければならない

#### Scenario: fmt.Errorfで%wを使わない
- **WHEN** 開発者が `fmt.Errorf` を使って上位へ返す `error` を生成するが、元 `err` を `%w` で連結していない
- **THEN** 品質ゲートは原因追跡不能な wrap 不足として報告しなければならない

#### Scenario: 同一変数への自己wrap再代入を行う
- **WHEN** 開発者が `err = fmt.Errorf("...: %w", err)` のように、同じ `error` 変数へ再代入しながら self wrap する
- **THEN** 品質ゲートは error message 肥大化リスクとして報告しなければならない

#### Scenario: cleanup以外で失敗を握りつぶす
- **WHEN** 開発者が cleanup 以外の本流処理で `error` を捨てる、または無視して正常系を継続する
- **THEN** 品質ゲートは error の握りつぶしとして報告しなければならない

### Requirement: error wrap検査は正当な例外を区別しなければならない
システムは、cleanup の best-effort 処理、`errors.Is` / `errors.As` 前提の明示的な変換、再送出不要なローカル補助処理など、文脈付き wrap を要求しない正当な例外を区別しなければならない。

#### Scenario: cleanup例外は本流違反と区別される
- **WHEN** 開発者が close / rollback / deferred cleanup の best-effort 失敗を記録専用で扱う
- **THEN** 品質ゲートは本流の wrap 不足と同じ重大度で誤報してはならない

#### Scenario: 明示的な error 変換は `%w` 必須対象から除外される
- **WHEN** 開発者が sentinel error への変換や `errors.Is` / `errors.As` 前提の分岐結果として、新しい `error` を返す
- **THEN** 品質ゲートは元 `error` を引数に保持しない変換まで `%w` 欠落として報告してはならない

#### Scenario: ローカル helper 再送出は MVP の必須境界に含めない
- **WHEN** 開発者が非公開 helper 内で受け取った `error` を同一 package の内部都合として返す
- **THEN** 品質ゲートは package 境界違反として一律に報告してはならない

### Requirement: 高コストな設計lintはblocking導入前に成立性評価を完了しなければならない
システムは、公開メソッドの責務過多のように誤検知コストが高い設計 lint をバックエンド品質ゲートへ追加する前に、実装可否、誤検知率、導入コスト、および最小 MVP を整理した成立性評価を完了しなければならない。

#### Scenario: 設計lint候補をblocking追加したい
- **WHEN** 開発者が公開メソッドの責務過多を検出する設計 lint を `backend:lint` の失敗条件へ加えようとする
- **THEN** システムは事前に MVP ルール、想定誤検知、導入コストを整理した成立性評価を持たなければならない

#### Scenario: 誤検知が高くblocking不適と判断される
- **WHEN** 成立性評価の結果、責務過多の判定が安定せず誤検知率が高いと分かった
- **THEN** システムは当該ルールを blocking lint として追加せず、レビュー観点または任意チェックとして扱う判断を残さなければならない

#### Scenario: MVPが成立すると判断される
- **WHEN** 成立性評価で、検出対象を限定した MVP なら実運用可能と判断された
- **THEN** システムは検出対象、除外境界、導入段階を定義した上で次 change に実装を進められなければならない
