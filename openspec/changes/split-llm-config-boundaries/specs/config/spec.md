## MODIFIED Requirements

### Requirement: config は保存技術と workflow 解釈を混在させてはならない
システムは、設定値の保存・migration と workflow 固有の default / 解釈を同一 package に混在させてはならない。設定保存は gateway 側の adapter から提供され、workflow 固有 default は workflow 側で解釈されなければならない。

#### Scenario: gateway が workflow config を再エクスポートしない
- **WHEN** 開発者が gateway 境界で config store を公開する
- **THEN** gateway は workflow 側 package を単純再エクスポートしてはならない
- **AND** gateway 自身の contract / implementation として store を公開しなければならない

#### Scenario: runtime が実行時設定を読み取る
- **WHEN** runtime が provider や model、endpoint、concurrency のような設定値を利用する
- **THEN** runtime は専用の読取補助を通じて設定を読み取れなければならない
- **AND** workflow 固有 default へ直接依存してはならない
