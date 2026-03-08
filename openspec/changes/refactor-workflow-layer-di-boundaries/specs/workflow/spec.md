## ADDED Requirements

### Requirement: Workflow は controller と usecase slice の接続を契約経由で担わなければならない
システムは `pkg/controller`、`pkg/workflow`、`pkg/usecase_slice`、`pkg/runtime`、`pkg/gateway` の責務区分を前提とし、controller から usecase slice および runtime への接続を `workflow` が集約して担わなければならない。controller は workflow の契約のみを呼び出し、usecase slice の具体的な実行順序、phase、resume、cancel、進捗決定を直接扱ってはならない。

#### Scenario: controller は workflow 契約だけを呼び出す
- **WHEN** UI が MasterPersona の開始要求を送信する
- **THEN** controller は workflow の開始契約を呼び出さなければならない
- **AND** controller は persona slice や queue の契約を直接呼び出してはならない

#### Scenario: workflow がユースケース進行を制御する
- **WHEN** MasterPersona の再開要求が送信される
- **THEN** workflow は parser、persona slice、runtime queue の契約を呼び分けてユースケース進行を制御しなければならない
- **AND** phase、resume、cancel、進捗の決定を workflow の責務として扱わなければならない

### Requirement: Workflow は slice 間の DTO マッピングを担わなければならない
workflow は、controller、parser、runtime、usecase slice の間で必要となる DTO 変換を担わなければならない。usecase slice は自前 DTO のみを契約として公開し、他 slice や外部入力の DTO に直接依存してはならない。

#### Scenario: parser 出力は workflow で persona 入力へ変換される
- **WHEN** workflow が parser から JSON 解析結果を受け取る
- **THEN** workflow は当該結果を persona slice の入力 DTO に変換しなければならない
- **AND** persona slice は parser の DTO を直接参照してはならない

#### Scenario: runtime 結果は workflow で保存用 DTO へ変換される
- **WHEN** workflow が runtime から完了済み request 結果を受け取る
- **THEN** workflow は persona slice の保存契約へ渡せる形に変換しなければならない
- **AND** persona slice は runtime 固有 DTO を理解してはならない

### Requirement: 全ての責務区分接続は DI インターフェース経由で行われなければならない
controller、workflow、usecase slice、runtime、gateway の接続は、各 package が公開する契約を通じて行われなければならない。具象型の配線は composition root に限定し、通常の package 実装が他区分の具象型へ直接依存してはならない。

#### Scenario: composition root だけが具象型を知る
- **WHEN** アプリケーション起動時に依存関係を組み立てる
- **THEN** composition root は具象型を生成して各契約へ配線しなければならない
- **AND** controller、workflow、usecase slice の通常 package は他区分の具象型を直接 import してはならない

#### Scenario: DI ルール違反は lint で検出される
- **WHEN** 開発者が区分間の具象依存や逆依存を追加する
- **THEN** `go-cleanarch` を用いた lint は当該違反を検出しなければならない
- **AND** 品質ゲートは違反を見逃してはならない
