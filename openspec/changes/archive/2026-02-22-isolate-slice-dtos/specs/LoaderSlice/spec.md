## MODIFIED Requirements

### Requirement: Loader出力の独自DTO定義とグローバルドメイン依存の排除
**Reason**: VSA（Vertical Slice Architecture）の原則に従い、システム全体で共有するデータモデル（`pkg/domain` 等）を排除し、各スライスが自身の入出力構造を独自定義する設計（Anti-Corruption Layer）へ移行するため。
**Migration**: これまでパース結果を格納するために利用していた `pkg/domain` などの外部依存モデルを破棄し、本スライスの `contract.go` パッケージ内に独自定義した出力専用DTOを返すようにインターフェースおよび内部実装を修正する。

#### Scenario: 独自定義DTOによるパース結果の返却
- **WHEN** Modファイル群のロードおよびパース処理が完了した場合
- **THEN** 外部パッケージのモデルに一切依存することなく、本スライス内部で定義された独自DTOに全データを格納して返却できること
