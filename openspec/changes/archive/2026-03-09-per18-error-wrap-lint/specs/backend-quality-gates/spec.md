## ADDED Requirements

### Requirement: error wrap不足を品質ゲートで検出しなければならない
システムは、package 境界または外部 I/O 境界で返却される `error` に文脈付き wrap が不足しているケースを、バックエンド品質ゲートで検出しなければならない。検査は、単純な `return err`、`fmt.Errorf` における `%w` 欠落、および cleanup 以外での失敗握りつぶしを対象に含めなければならない。

- package 境界には、公開関数、公開メソッド、および他 package から利用される公開 Contract 実装の返却点を含めなければならない。
- 外部 I/O 境界には、gateway / datastore / 外部 API / ファイル I/O の失敗を上位へ返す箇所を含めなければならない。
- `fmt.Errorf` 検査は、元 `error` を引数に含みつつ `%w` が使われていないケースに限定しなければならない。
- 品質ゲートは、同じ `error` 変数に対する `err = fmt.Errorf(\"...: %w\", err)` の自己 wrap 再代入を検出しなければならない。

#### Scenario: package境界でreturn errを返す
- **WHEN** 開発者が package 境界の公開メソッドまたは外部依頼の呼び出し箇所で、取得した `err` を文脈付与せず `return err` する
- **THEN** 品質ゲートは error wrap 不足として報告しなければならない

#### Scenario: fmt.Errorfで%wを使わない
- **WHEN** 開発者が `fmt.Errorf` を使って上位へ返す `error` を生成するが、元 `err` を `%w` で連結していない
- **THEN** 品質ゲートは原因追跡不能な wrap 不足として報告しなければならない

#### Scenario: 同一変数への自己wrap再代入を行う
- **WHEN** 開発者が `err = fmt.Errorf(\"...: %w\", err)` のように、同じ `error` 変数へ再代入しながら self wrap する
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
