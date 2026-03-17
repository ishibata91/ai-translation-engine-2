## ADDED Requirements

### Requirement: Translation Flow workflow は単語翻訳をデータロード直後の phase として扱わなければならない
システムは、`TranslationFlow` の workflow において `単語翻訳` を `データロード` 直後かつ `本文翻訳` より前の phase として扱わなければならない。workflow はロード済み入力が存在しない状態で単語翻訳 phase を開始してはならない。

#### Scenario: phase 順序を決定する
- **WHEN** workflow が translation flow の phase 順序を構築する
- **THEN** `単語翻訳` は `データロード` の直後に配置されなければならない
- **AND** ロード済みファイルが 0 件の場合は `単語翻訳` を開始不可として扱わなければならない

### Requirement: Translation Flow workflow は terminology slice に artifact 識別子と phase 設定 DTO だけを渡さなければならない
システムは、translation flow の単語翻訳 phase において、workflow から terminology slice へ渡す情報を task ID や file 識別子などの artifact 境界情報と、operator が確定した terminology 用 request/prompt 設定 DTO に限定しなければならない。workflow は artifact を直接読み出して terminology 用入力 DTO を組み立ててはならない。

#### Scenario: terminology phase を開始する
- **WHEN** workflow が単語翻訳 phase を開始する
- **THEN** workflow は task ID または同等の artifact 識別子と terminology 用設定 DTO を terminology slice に渡さなければならない
- **AND** workflow から artifact repository を直接呼び出してはならない

### Requirement: terminology slice は artifact から単語翻訳対象を抽出しなければならない
システムは、terminology slice が artifact から terminology 対象レコードだけを抽出し、slice 内で `TermTranslatorInput` を構築しなければならない。本文翻訳専用の本文や長文レコードを terminology 入力へ混在させてはならない。

#### Scenario: 単語翻訳対象を抽出する
- **WHEN** terminology slice が task ID を受け取って単語翻訳 phase を開始する
- **THEN** terminology の対象レコードタイプに含まれる NPC、Item、Magic、Message、Location、Quest の単語翻訳対象だけを artifact から抽出しなければならない
- **AND** 本文翻訳専用のレコード本文を terminology 入力へ混在させてはならない

### Requirement: Translation Flow workflow は terminology の 2 フェーズモデルを通じて単語翻訳を実行しなければならない
システムは、translation flow の単語翻訳 phase において terminology slice の `PreparePrompts` と `SaveResults` を順に利用し、Pass 1 の単語翻訳を完了しなければならない。workflow は terminology の永続化責務を自前実装してはならない。

#### Scenario: 単語翻訳ジョブを生成して保存する
- **WHEN** workflow が terminology slice を起動する
- **THEN** terminology slice は operator 指定設定を反映した LLM リクエスト群を返さなければならない
- **AND** workflow は対応するレスポンスを terminology slice の保存フェーズへ渡さなければならない

### Requirement: Translation Flow workflow は terminology 側の保存状態を phase 完了判定の正本として扱わなければならない
システムは、単語翻訳 phase の完了状態を task metadata ではなく terminology 側の保存状態から判定しなければならない。workflow は保存状態を JSON メタデータへ複製して正本化してはならない。

#### Scenario: 単語翻訳 phase の完了を判定する
- **WHEN** workflow が task の単語翻訳 phase 状態を取得する
- **THEN** terminology 側の status カラムに基づいて完了・進行中・失敗を判定しなければならない
- **AND** task metadata の JSON 値だけで完了判定してはならない

### Requirement: Translation Flow workflow は単語翻訳結果を本文翻訳前提の成果物として再利用できなければならない
システムは、単語翻訳 phase で保存された terminology の結果を本文翻訳 phase が同じ task の前提成果物として参照できるようにしなければならない。

#### Scenario: 本文翻訳へ handoff する
- **WHEN** 単語翻訳 phase が完了して本文翻訳 phase へ進む
- **THEN** workflow は本文翻訳 phase が保存済み用語辞書を参照できる状態を保証しなければならない
- **AND** 本文翻訳 phase で単語翻訳を再実行しなくてもよい状態にしなければならない

#### Scenario: 既存 task を再表示する
- **WHEN** ユーザーが単語翻訳済みの translation flow task を開き直す
- **THEN** workflow は terminology 側の保存状態から単語翻訳結果の存在を復元できなければならない
- **AND** 単語翻訳 phase の再実行を毎回必須にしてはならない

### Requirement: Translation Flow workflow は単語翻訳の実行サマリを返さなければならない
システムは、単語翻訳 phase 完了時に `対象件数 / 保存件数 / 失敗件数` を返し、UI と後続 phase が同じサマリを参照できるようにしなければならない。

#### Scenario: 単語翻訳結果を集計する
- **WHEN** workflow が単語翻訳 phase を完了する
- **THEN** 対象件数と保存件数を返さなければならない
- **AND** 失敗件数を 0 件と区別できる形で返さなければならない
