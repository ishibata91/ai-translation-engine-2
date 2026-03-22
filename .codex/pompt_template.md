件数が10件から先に進まない
ページングが機能してない

[$fix-direction](F:\ai translation engine 2\.codex\skills\fix-direction\SKILL.md)

## 不具合概要
- **画面/機能**: 翻訳プロジェクト/単語翻訳フェーズ
- **現象**: Gemini同期実行でエラー

## 再現手順
前提: データロード済み
1. 単語翻訳画面に遷移する
2. Gemini 同期実行をモデルに指定して翻訳実行

再現率: 毎回

## 期待挙動
全ての単語が翻訳できること 
## 実際の挙動
12件のうち､2件成功して他10件は成功しない
## 補足
- ログ: "non-retryable error on attempt 1: gemini: API error 404: "}
{"time":"2026-03-22T18:34:41.6820319+09:00","level":"DEBUG","msg":"action: llm.request, 45ms","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","action":"llm.request","duration_ms":44.617}
{"time":"2026-03-22T18:34:41.6820319+09:00","level":"WARN","msg":"EXIT executeOne: request failed","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","index":0,"error":"gemini: complete request failed: non-retryable error on attempt 1: gemini: API error 404: "}
{"time":"2026-03-22T18:34:41.6820319+09:00","level":"DEBUG","msg":"ENTER executeOne","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","index":1}
{"time":"2026-03-22T18:34:41.6820319+09:00","level":"DEBUG","msg":"action: llm.request","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","action":"llm.request"}
{"time":"2026-03-22T18:34:41.682548+09:00","level":"DEBUG","msg":"Gemini request start","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","system_prompt_len":1472,"user_prompt_len":28}
{"time":"2026-03-22T18:34:41.682548+09:00","level":"DEBUG","msg":"ENTER buildRequest","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","model":"models/gemini-3.1-flash-lite-preview"}
{"time":"2026-03-22T18:34:41.682548+09:00","level":"DEBUG","msg":"EXIT buildRequest","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","url_path":"/models/models/gemini-3.1-flash-lite-preview:generateContent"}
{"time":"2026-03-22T18:34:41.8748043+09:00","level":"ERROR","msg":"Gemini request failed","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","error_code":"UNKNOWN","exception_class":"wrapError","stack_trace":"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.(*geminiClient).Complete\n\tF:/ai translation engine 2/pkg/gateway/llm/gemini_client.go:114\ngithub.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.executeOne\n\tF:/ai translation engine 2/pkg/gateway/llm/bulk.go:192\ngithub.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.runWorkerPoolWithProgress.func1\n\tF:/ai translation engine 2/pkg/gateway/llm/bulk.go:156\nruntime.goexit\n\tC:/Program Files/Go/src/runtime/asm_amd64.s:1693\n","error_message":"non-retryable error on attempt 1: gemini: API error 404: "}
{"time":"2026-03-22T18:34:41.8748043+09:00","level":"DEBUG","msg":"action: llm.request, 193ms","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","action":"llm.request","duration_ms":192.772}
{"time":"2026-03-22T18:34:41.8748043+09:00","level":"WARN","msg":"EXIT executeOne: request failed","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","index":1,"error":"gemini: complete request failed: non-retryable error on attempt 1: gemini: API error 404: "}
{"time":"2026-03-22T18:34:41.8748043+09:00","level":"DEBUG","msg":"ENTER executeOne","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","index":2}
{"time":"2026-03-22T18:34:41.8755396+09:00","level":"DEBUG","msg":"action: llm.request","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","action":"llm.request"}
{"time":"2026-03-22T18:34:41.8755396+09:00","level":"DEBUG","msg":"Gemini request start","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","system_prompt_len":756,"user_prompt_len":28}
{"time":"2026-03-22T18:34:41.8755396+09:00","level":"DEBUG","msg":"ENTER buildRequest","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","model":"models/gemini-3.1-flash-lite-preview"}
{"time":"2026-03-22T18:34:41.8755396+09:00","level":"DEBUG","msg":"EXIT buildRequest","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","url_path":"/models/models/gemini-3.1-flash-lite-preview:generateContent"}
{"time":"2026-03-22T18:34:41.920278+09:00","level":"ERROR","msg":"Gemini request failed","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","env":"development","app_version":"unknown","service_name":"ai-translation-engine","host_name":"DESKTOP-F82R7TM","component":"llm_manager","component":"gemini_client","model":"models/gemini-3.1-flash-lite-preview","error_code":"UNKNOWN","exception_class":"wrapError","stack_trace":"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.(*geminiClient).Complete\n\tF:/ai translation engine 2/pkg/gateway/llm/gemini_client.go:114\ngithub.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.executeOne\n\tF:/ai translation engine 2/pkg/gateway/llm/bulk.go:192\ngithub.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm.runWorkerPoolWithProgress.func1\n\tF:/ai translation engine 2/pkg/gateway/llm/bulk.go:156\nruntime.goexit\n\tC:/Program Files/Go/src/runtime/asm_amd64.s:1693\n","error_message":"non-retryable error on attempt 1: gemini: API error 404: "}
- 波及の可能性:なし 
- エラー表示: なし
- スクリーンショット:なし 

---

[$plan-direction](F:\ai translation engine 2\.codex\skills\plan-direction\SKILL.md)

## 設計・仕様策定の概要
- **対象機能/画面**: 翻訳プロジェクト/ペルソナ生成フェーズ
- **目的・背景**: 翻訳対象に登場するNPCの口調などの､ペルソナを生成する｡

## 要求事項
1. 翻訳対象に登場するNPCのペルソナを生成できること
2. 既存のマスターペルソナに含まれているNPCは生成対象としないこと


## 制約・前提条件


## 補足資料
- 関連資料/issue: 
- docs\slice\persona 既存仕様
- 

---

[$impl-direction](F:\ai translation engine 2\.codex\skills\impl-direction\SKILL.md)

## 実装対象の概要
- **対象の変更文書 (changes)**: translation-flow-terminology-dictionary-reference-rules
- **目的**: ペルソナ生成フェーズを実装する

## 実装スコープ
- **対象領域**: [Frontend / Backend / Mixed]
Mixed
## 前提・制約事項
- **共有コントラクト**: 
- **所有範囲 (owned paths)**: 
- **変更禁止範囲 (forbidden paths)**: 

## 補足・懸念事項
- 
