# 設定・レイアウト保存インフラ (Config Store) 仕様書

## 概要
アプリケーション全体の設定情報（APIキー、LLM選択状態、UIレイアウト等）をSQLiteで永続化するインフラストラクチャ層の設計である。
Interface-First AIDD v2 アーキテクチャに則り、各Sliceや上位層が共通で利用できる**横断的インフラモジュール**として設計する。

## 背景・動機
- 現状、APIキーは環境変数やハードコードで管理されており、UIからの動的な変更ができない。
- LLMプロバイダーの選択状態やUIレイアウト（ウィンドウサイズ、パネル配置等）を永続化する仕組みがない。
- 各Sliceが独自に設定管理を実装すると重複が生じるため、共通インフラとして一元管理する。

## 設計方針

### テーブル分離の原則（1:1分離）
設定データを**用途と型安全性の要件**に基づき、以下の3テーブルに分離する。

| テーブル | 用途 | 値の形式 | 利用者 |
| :--- | :--- | :--- | :--- |
| `config` | バックエンド設定 | **型付きTEXT**（プレーン文字列。JSONは許可しない） | Go バックエンド |
| `ui_state` | UIレイアウト・UI状態 | **JSON許可**（構造化データを自由に格納） | React フロントエンド |
| `secrets` | 機密情報 | **プレーンTEXT**（将来暗号化対応） | Go バックエンド |

**分離の理由**:
- **バックエンド設定 (`config`)**: LLMプロバイダー名、モデル名、temperature等はGoコードが直接読み取る。JSON文字列をパースする必要がなく、`string`/`int`/`float`/`bool`として直接扱えるべきである。値は常にプレーンな文字列表現（例: `"gemini"`, `"0.1"`, `"true"`）で格納し、`TypedAccessor`が型変換を担う。
- **UIステート (`ui_state`)**: パネルサイズ配列 `[300,500,200]` やカラム設定等、構造が可変でフロントエンド固有のデータはJSON形式での格納が合理的。バックエンドはこのデータを解釈せず、透過的に保存・返却するだけである。
- **シークレット (`secrets`)**: APIキー等の機密情報は独立テーブルで管理し、暗号化・アクセス制御の拡張ポイントとする。

### その他の方針
1. **Key-Valueストア**: 各テーブルとも `namespace + key → value` の形式で格納する。名前空間により各Sliceやコンテキストの設定を論理的に分離する。
2. **型安全なアクセス**: `config`テーブルに対しては型付きヘルパー関数（`GetString`, `GetInt`, `GetFloat`, `GetBool`）を提供する。`GetJSON`/`SetJSON`は`ui_state`テーブル専用とする。
3. **リアクティブ通知**: 設定変更時にコールバック/チャネルで通知する仕組みを持ち、UIやプロバイダーの動的切り替えを可能にする。
4. **SQLite単一ファイル**: アプリケーションデータディレクトリ内の単一 `.db` ファイルで管理し、ポータビリティを確保する。

## 要件

### 機能要件
1. **バックエンド設定の読み書き**: `config`テーブルに対して `namespace/key` でプレーン文字列値を保存・取得・削除できる。JSON値の格納は許可しない。
2. **UIステートの読み書き**: `ui_state`テーブルに対して `namespace/key` でJSON値を保存・取得・削除できる。
3. **APIキー管理**: UIからLLMプロバイダーごとのAPIキーを登録・更新・削除できる。APIキーは `secrets` テーブルに格納する。
4. **LLM選択状態の保存**: ユーザーが最後に選択したLLMプロバイダー・モデル・パラメータを `config` テーブルに永続化し、次回起動時に復元する。
5. **UIレイアウトの保存**: ウィンドウサイズ、パネル配置、表示カラム等のUI状態を `ui_state` テーブルに永続化する。
6. **名前空間による分離**: 各Slice（`dictionary_builder`, `translator` 等）やシステム（`ui`, `llm`）ごとに名前空間を持つ。
7. **デフォルト値**: キーが未設定の場合にデフォルト値を返す仕組みを提供する。
8. **変更通知**: 特定の名前空間/キーの変更を監視し、コールバックで通知する。

### 非機能要件
1. **スレッドセーフ**: 複数のGoroutineから同時にアクセスしても安全であること。
2. **起動時マイグレーション**: スキーマバージョンを管理し、アプリ起動時に自動マイグレーションを行う。
3. **テスタビリティ**: インターフェース経由でアクセスし、テスト時はインメモリSQLiteに差し替え可能とする。

## データモデル

### テーブル: `config`（バックエンド設定）
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `namespace` | TEXT NOT NULL | 名前空間 (例: `llm`, `dictionary_builder`) |
| `key` | TEXT NOT NULL | 設定キー (例: `selected_provider`, `temperature`) |
| `value` | TEXT | プレーン文字列値（JSON不許可。例: `gemini`, `0.1`, `true`） |
| `updated_at` | DATETIME | 最終更新日時 |
| PRIMARY KEY | | `(namespace, key)` |

### テーブル: `ui_state`（UIステート）
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `namespace` | TEXT NOT NULL | 名前空間 (例: `ui.layout`, `ui.preferences`) |
| `key` | TEXT NOT NULL | 設定キー (例: `panel_sizes`, `column_config`) |
| `value` | TEXT | JSON形式の値（構造化データ許可） |
| `updated_at` | DATETIME | 最終更新日時 |
| PRIMARY KEY | | `(namespace, key)` |

### テーブル: `secrets`
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `namespace` | TEXT NOT NULL | 名前空間 (例: `llm.gemini`, `llm.openai`) |
| `key` | TEXT NOT NULL | シークレットキー (例: `api_key`) |
| `value` | TEXT | 値（将来的に暗号化対応） |
| `updated_at` | DATETIME | 最終更新日時 |
| PRIMARY KEY | | `(namespace, key)` |

### テーブル: `schema_version`
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `version` | INTEGER | 現在のスキーマバージョン |
| `applied_at` | DATETIME | 適用日時 |

## 名前空間の規約

### `config` テーブル（バックエンド設定）
| 名前空間 | 用途 | キー例 | 値の例 |
| :--- | :--- | :--- | :--- |
| `llm` | LLM共通設定 | `selected_provider`, `selected_model`, `temperature`, `max_tokens` | `gemini`, `gemini-pro`, `0.1`, `4096` |
| `llm.gemini` | Geminiプロバイダー固有 | `endpoint` | `https://generativelanguage.googleapis.com` |
| `llm.openai` | OpenAIプロバイダー固有 | `endpoint` | `https://api.openai.com` |
| `llm.local` | ローカルLLM固有 | `server_url`, `model_path` | `http://localhost:1234`, `/models/gemma` |

### `ui_state` テーブル（UIステート）
| 名前空間 | 用途 | キー例 | 値の例 |
| :--- | :--- | :--- | :--- |
| `ui.layout` | UIレイアウト | `window_size`, `panel_sizes`, `sidebar_visible` | `{"w":1200,"h":800}`, `[300,500,200]`, `true` |
| `ui.preferences` | UI設定 | `theme`, `language`, `font_size` | `"dark"`, `"ja"`, `14` |

### `secrets` テーブル（機密情報）
| 名前空間 | 用途 | キー例 |
| :--- | :--- | :--- |
| `llm.gemini` | Gemini APIキー | `api_key` |
| `llm.openai` | OpenAI APIキー | `api_key` |
| `llm.xai` | xAI APIキー | `api_key` |

## ライブラリの選定
- DBアクセス: `database/sql` + `github.com/mattn/go-sqlite3`
- 依存性注入: `github.com/google/wire`
- JSON操作: Go標準 `encoding/json`（`ui_state`テーブル用のみ）

## 関連ドキュメント
- [クラス図](./config_store_class_diagram.md)
- [シーケンス図](./config_store_sequence_diagram.md)
- [テスト設計](./config_store_test_spec.md)
