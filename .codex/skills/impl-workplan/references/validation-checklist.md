# Validation Checklist

この section の正本 command は次です。

```powershell
powershell -ExecutionPolicy Bypass -File .codex/skills/impl-workplan/scripts/validate-skill-contracts.ps1
```

exit code `0` は contract 崩れなし、`1` は review 前に解消すべき崩れありを意味する。

## Hotfix

- `SKILL.md` `references/*.md` `agents/*.toml` を触った直後に同じ validator command を再実行する。
- missing reference が 0 件で、追加した reference/template/script 参照がすべて実在することを確認する。
- deprecated tool name が 0 件で、旧 tool 名を再導入していないことを確認する。
- fix logging 関連の記述と agent 指示が `[fix-trace]` prefix に揃っていることを確認する。

## Workflow Normalization

- impl lane の section schema が `validation_commands` を含む正本 field 名で揃っていることを確認する。
- `changes/<id>/tasks.md` の各 section に `validation_commands` があり、legacy field (`validation` `validation_command` `quality_gates`) が残っていないことを確認する。
- nested section schema (`sections:` `affected_sections:` 配下) でも `validation_commands:` が false positive なしで認識されることを確認する。
- direction / workplan / worker の文書で、同じ work order field 名を参照していることを確認する。
- workflow 正規化後も同じ validator command だけで drift を再検出できることを確認する。

## Validation Automation

- review 前の必須 gate として validator command を 1 回実行し、結果を handoff に残す。
- Hotfix 後と Workflow Normalization 後に同じ command を再実行し、再監査に追加スクリプトが不要であることを確認する。
- 新しい skill / agent / tasks.md を追加した場合でも、validator の再帰走査だけで対象に自動で含まれることを確認する。
- validator の失敗は放置せず、finding を解消するか、所有外なら blocking issue として明示する。
