## MODIFIED Requirements

### Requirement: progress は横断基盤として利用できなければならない
システムは、progress notifier と event transport を runtime 固有ではなく横断基盤として提供し、`controller`、`workflow`、`slice`、`runtime`、`gateway` から利用できなければならない。

#### Scenario: LLM 周辺が foundation progress を利用する
- **WHEN** LLM 周辺の workflow または runtime が進捗通知を送る
- **THEN** progress は foundation 配下の notifier を通じて利用できなければならない
- **AND** progress が workflow の phase 解釈を内包してはならない
