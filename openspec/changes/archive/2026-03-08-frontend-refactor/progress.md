# Progress Log

## 2026-03-08

- 規約文書を整備し、例外許容箇所を「なし」で固定。
- ESLint / Prettier / TypeScript の品質ゲートを導入。
- import 境界ルール（`pages -> wailsjs/store` 禁止）を導入。
- `DictionaryBuilder` を `state/actions/ui/constants` 契約へ再編。
- Wailsレスポンス変換を `adapters.ts` へ抽出。
- `useWailsEvent` 共通Hookを導入し、イベント購読処理を共通化。
- `frontend/src` 配下の `any` を `unknown` ベースへ置換。
- `useMasterPersona` のユーティリティ責務を `helpers.ts` へ分離。
- `vitest` + Testing Library + `msw` を導入し、`useDictionaryBuilder` の振る舞いテストを追加。
- Wails 正式コマンド (`wails build`) で `frontend/src/wailsjs` を再生成。
- 品質ゲート実行結果:
  - `npm run typecheck`: pass
  - `npm run lint`: pass (warning あり)
  - `npm run test`: pass
  - `npm run build`: pass
