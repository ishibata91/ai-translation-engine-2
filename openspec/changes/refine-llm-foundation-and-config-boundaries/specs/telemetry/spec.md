## MODIFIED Requirements

### Requirement: telemetry は横断基盤として利用できなければならない
システムは、logger provider、context 補助、span 補助、Wails bridge を runtime 固有ではなく横断基盤として提供し、`controller`、`workflow`、`slice`、`runtime`、`gateway` から利用できなければならない。

#### Scenario: gateway が foundation telemetry を利用する
- **WHEN** LLM gateway が request span や error attribute を記録する
- **THEN** gateway は foundation 配下の telemetry を利用できなければならない
- **AND** gateway が runtime へ直接依存しなくても同じ観測情報を出力できなければならない
