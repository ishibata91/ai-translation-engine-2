# xTranslator XML (SSTXMLRessources) 仕様メモ

xTranslator が出力・使用する XML ベースの辞書/翻訳リソースファイルについての仕様メモです。
スカイリムなどのゲームMod翻訳において、文字列のインポート/エクスポートファイルとして利用されます。

## 全体構造

XMLのルート要素は `<SSTXMLRessources>` であり、配下にメタデータを表す `<Params>` と、実際の翻訳データを格納する `<Content>` ブロックが存在します。

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<SSTXMLRessources>
  <Params>
    <!-- メタデータ -->
  </Params>
  <Content>
    <!-- 翻訳データ要素群 -->
  </Content>
</SSTXMLRessources>
```

---

## 1. Params 要素 (メタデータ)

辞書や翻訳対象プラグインの基本情報が指定されます。

*   **`<Addon>`**: どのプラグインやModに属する文字列かを示します。(例: `Dawnguard`)
*   **`<Source>`**: 元言語を示す文字列です。(通常 `english`)
*   **`<Dest>`**: 翻訳先言語を示す文字列です。(例: `japanese`)
*   **`<Version>`**: XMLフォーマットのバージョン情報です。(通常 `2`)

---

## 2. Content 要素 (翻訳データリスト)

各用語・翻訳テキストのペアは、`<Content>` 配下の `<String>` 要素として列挙されます。

```xml
<String List="0" sID="000001" Partial="1">
  <EDID>DLC1AurielsBow</EDID>
  <REC>WEAP:FULL</REC>
  <Source>Auriel's Bow</Source>
  <Dest>アーリエルの弓</Dest>
</String>
```

### `<String>` の属性
*   `List="0"`: 通常は 0 が指定されます。コンテキストリストの種類を表すと考えられます。
*   `sID="XXXXXX"`: ストリングID。プラグイン内での識別子（FormIDの下位バイト等と関連）ですが、辞書利用時には必ずしも一意に信用できるキーではありません。
*   `Partial="1"`: 部分一致やステータスを示すフラグの可能性があります（多くの場合 `1` が設定されています）。

### `<String>` 配下の子要素
辞書DBの構築において、以下の4要素が最も重要です。

*   **`<EDID>` (Editor ID)**
    *   Creation Kit などにおける Editor ID (オブジェクトの人間可読な識別名) です。
    *   例: `DLC1AurielsBow`, `DLC1VampireBeastRace`
    *   **キーとして重要**: 翻訳を適用する際、フォームIDが変わってもEDIDの一致で翻訳を特定できるため、非常に強力な識別子となります。
*   **`<REC>` (Record Type : Signature)**
    *   対象テキストが属するレコードの種類と、フィールド(シグネチャ)の組み合わせを示します。
    *   例:
    *   例:
        *   `WEAP:FULL`: 武器(WEAP)のフルネーム(FULL)
        *   `NPC_:FULL`: NPCのフルネーム
        *   `MGEF:DNAM`: 魔法効果(MGEF)の説明テキスト(DNAM)
        *   `DIAL:FULL`: ダイアログ(会話テキスト)
        *   `QUST:FULL`: クエスト名や目標
    *   **複合キー**: 用語の特定には `EDID` と `REC` の組み合わせを使用するのが確実です。
*   **`<Source>`**
    *   原文（英語など）のテキストです。
    *   CDO(CDATA)で囲まれることは少なく、エスケープ文字(`&apos;` など)が使用されます。空の場合は空タグ `<Source />` になります。
*   **`<Dest>`**
    *   翻訳文（日本語など）のテキストです。
    *   空の場合は空タグになります。

---

## xTranslator XML の一般的な用途

この XML フォーマットは、Skyrim や Fallout などの Bethesda 製ゲームエンジンにおいて、各言語の .STRINGS ファイルやマスターファイル（.esm, .esp）から翻訳データを外部ファイルとしてエクスポートし、xTranslator エディタ上で翻訳作業を行うために特化しています。
