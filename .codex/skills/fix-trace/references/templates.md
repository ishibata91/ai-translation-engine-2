# Debugger Templates

正本は `changes/<id>/context_board/fix-trace.packet.json` とし、validation は同じ場所の
`fix-trace.packet.validation.json` を使う。

## trace packet
```json
{
  "invoked_skill": "fix-trace",
  "invoked_by": "fix-direction",
  "change": "<change-id>",
  "current_hypothesis": "",
  "unknowns": [],
  "current_scope": [],
  "next_action": "",
  "cause_hypothesis": {},
  "logger_plan": {},
  "debugger_findings": {}
}
```

## cause hypothesis
```md
### Cause Hypothesis
- symptom:
- likely layers:
- primary hypothesis:
- alternative hypotheses:
- current_scope:
- unknowns:
- next observation:
```

## logger plan
```md
### Logger Plan
- target files:
- logger path:
- output files:
- summary_update:
- optional stack trace:
```

## debugger findings
```md
### Debugger Findings
- confirmed facts:
- narrowed cause:
- remaining uncertainty:
- next_action:
- suggested fix scope:
```

## Rules

- `invoked_skill` `invoked_by` `change` `current_hypothesis` `unknowns` `current_scope` `next_action` は validator の最低必須 field として毎回出力する。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
