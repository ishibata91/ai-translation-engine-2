## 1. 文書棚卸し

- [ ] 1.1 `openspec/specs` 配下を棚卸しし、UI / workflow / runtime / gateway の共通要件がユースケース spec に混在している箇所を列挙する
- [ ] 1.2 `architecture.md` と他 spec の重複箇所を洗い出し、移管先を決める

## 2. アーキテクチャ文書の整理

- [ ] 2.1 `openspec/specs/architecture.md` を純粋な構造文書として再構成する
- [ ] 2.2 品質ゲート、テスト設計、ログ、フロント構造への参照境界を明記する

## 3. spec 構成の再編

- [ ] 3.1 共通要件を置く spec の分割方針を定義する
- [ ] 3.2 必要な共通 spec を新設または既存 spec を再編する
- [ ] 3.3 `AGENTS.md` の参照ルールを整理後の spec 構成へ合わせて更新する
