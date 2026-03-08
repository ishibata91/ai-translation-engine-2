## 1. controller / workflow 接続

- [ ] 1.1 MasterPersona の開始・再開・キャンセルを扱う controller adapter を用意する
- [ ] 1.2 workflow 側に MasterPersona 用の開始・再開・キャンセル入口を実装する
- [ ] 1.3 `pkg/task` に残る controller 相当責務を `pkg/controller` へ移すか、暫定 adapter として責務を固定する

## 2. 実行経路の移行

- [ ] 2.1 開始経路を `controller -> workflow -> persona / runtime` に移す
- [ ] 2.2 resume / cancel / progress / phase 更新を workflow 主導へ移す
- [ ] 2.3 runtime 結果 -> persona 保存 DTO のマッピングを workflow へ集約する

## 3. 互換性と検証

- [ ] 3.1 既存 task API のシグネチャを維持したまま内部接続先を workflow へ差し替える
- [ ] 3.2 MasterPersona の開始・再開・キャンセル・cleanup のテストを更新する
- [ ] 3.3 `backend:lint:file` と `lint:backend` で最終確認する
- [ ] 3.4 `go-cleanarch` の controller / workflow 区分が実コード配置を反映するよう更新する
