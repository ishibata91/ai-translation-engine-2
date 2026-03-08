# Purpose

controller から usecase slice と runtime を接続する workflow capability の責務と要件を定義する。

## Requirements

### Requirement: Workflow は controller の唯一のユースケース入口でなければならない
controller は usecase slice や runtime を直接呼び出さず、workflow 契約だけを呼び出さなければならない。

#### Scenario: controller は workflow だけを呼び出す
- **WHEN** UI がユースケース開始または再開を要求する
- **THEN** controller は workflow 契約だけを呼び出さなければならない

### Requirement: Workflow は runtime を通じて実行制御を行わなければならない
workflow は queue、progress、workflow state などの実行制御を runtime 契約経由で行わなければならない。

#### Scenario: workflow は runtime queue を利用する
- **WHEN** workflow が request を enqueue または resume する
- **THEN** workflow は runtime queue 契約を利用しなければならない
