# 境界基盤

### Requirement: foundation は横断基盤専用の責務区分でなければならない
システムは、`telemetry` と `progress` のように複数責務区分から利用される横断基盤を `foundation` 区分として扱わなければならない。foundation は `controller`、`workflow`、`slice`、`runtime`、`gateway` から参照できなければならない。

#### Scenario: 横断基盤を foundation に配置する
- **WHEN** 開発者が `telemetry` や `progress` のような横断基盤を追加または移設する
- **THEN** 当該基盤は `foundation` 区分に配置されなければならない
- **AND** runtime 固有基盤として扱ってはならない

### Requirement: foundation は業務進行の意味づけを持ってはならない
システムは、foundation 配下に transport / observability の補助のみを置き、phase や progress 比率の意味づけ、workflow 状態解釈を含めてはならない。

#### Scenario: progress notifier を foundation に置く
- **WHEN** 開発者が foundation 配下の progress notifier を実装する
- **THEN** notifier は通知 transport だけを提供しなければならない
- **AND** 何を何%として通知するかの解釈は workflow 側に残さなければならない
