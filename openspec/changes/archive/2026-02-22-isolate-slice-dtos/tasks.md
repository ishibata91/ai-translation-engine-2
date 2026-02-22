## 1. TermTranslatorSlice

- [x] 1.1 `pkg/term_translator/contract.go` に独自入力DTO（例: `TermTranslatorInput`）を定義する
- [x] 1.2 `TermTranslatorSlice` の各メソッド引数を、外部DTOから独自DTOに変更する
- [x] 1.3 内部の処理ロジックおよびテストコードを独自DTOを使用するように修正する

## 2. PersonaGenSlice

- [x] 2.1 `pkg/persona_gen/contract.go` に独自入力DTO（例: `PersonaGenInput`）を定義する
- [x] 2.2 `PersonaGenSlice` の各メソッド引数を、外部DTOから独自DTOに変更する (実装なしのためスキップ)
- [x] 2.3 内部の処理ロジックおよびテストコードを独自DTOを使用するように修正する (実装なしのためスキップ)

## 3. ContextEngineSlice

- [x] 3.1 `pkg/context_engine/contract.go` に独自入力DTOを定義する
- [x] 3.2 `ContextEngineSlice` の各メソッド引数を変更する (実装なしのためスキップ)
- [x] 3.3 内部の処理ロジックおよびテストコードを修正する (実装なしのためスキップ)

## 4. SummaryGeneratorSlice

- [x] 4.1 `pkg/summary_generator/contract.go` に独自入力DTOを定義する (対応済)
- [x] 4.2 `SummaryGeneratorSlice` の各メソッド引数を変更する (対応済)
- [x] 4.3 内部の処理ロジックおよびテストコードを修正する (対応済)

## 5. Pass2TranslatorSlice

- [x] 5.1 `pkg/pass2_translator/contract.go` に独自入力DTOを定義する
- [x] 5.2 `Pass2TranslatorSlice` の各メソッド引数を、外部DTOから独自DTOに変更する (実装なしのためスキップ)
- [x] 5.3 内部の処理ロジックおよびテストコードを独自DTOを使用するように修正する (実装なしのためスキップ)

## 6. LoaderSlice

- [x] 6.1 `pkg/loader_slice/contract.go` に自スライス専用の出力DTO（例: パース結果モデル）を定義する
- [x] 6.2 `pkg/domain` への依存を排除し、独自の出力DTOを返すように各メソッドのシグネチャと処理を修正する
- [x] 6.3 関連するテストコードを修正する

## 7. Cleanup

- [x] 7.1 グローバルな共有モデルである `pkg/domain` フォルダ全体を削除する
- [x] 7.2 不要になった旧DTOや他スライスへの依存コードの残骸をクリーンアップする

## 8. ProcessManager (Orchestrator - 準備のみ)

- [x] 8.1 `specs/ProcessManagerSlice/spec.md` を更新または確認し、将来的に各スライスのDTOへマッピングを行う責務について明記する
