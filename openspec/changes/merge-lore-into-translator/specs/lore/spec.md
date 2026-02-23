## REMOVED Requirements

### 1. 会話ツリー解析
**Reason**: `translator` スライスに統合され、内部コンポーネントとして再定義されたため。
**Migration**: `translator` スライスが直接ゲームデータを受け取り、内部で解析を行う。

### 2. 話者プロファイリング
**Reason**: `translator` スライスに統合されたため。
**Migration**: `translator` スライス内部でペルソナDBやNPC属性を参照する。

### 3. 参照用語検索
**Reason**: `translator` スライスに統合されたため。
**Migration**: `translator` スライス内部で辞書検索を実行する。

### 4. 翻訳リクエストの構築
**Reason**: `TranslationRequest` というスライス間の中間DTOを廃止し、`llm.Request` を直接生成するフローに刷新したため。
**Migration**: 統合された `translator` スライス（Pass 2 Translator）がプロセスマネージャーから直接入力を受け取り、ジョブを提案する。

### 5. 要約の参照
**Reason**: `translator` スライスに統合されたため。
**Migration**: `translator` スライス内部で要約データを参照する。
