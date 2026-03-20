# Explorer Response Template

explorer subagent の最終返答は XML 形式で返す。
期待挙動、推測、原因候補、評価語は禁止し、事実のみを書く。

## Template

```xml
<context_packet>
  <scope>
    <target></target>
    <requested_scan_range>
      <path></path>
    </requested_scan_range>
    <scanned_paths>
      <path></path>
    </scanned_paths>
    <unscanned_paths>
      <path reason=""></path>
    </unscanned_paths>
  </scope>
  <reading_guide>
    <how_to_read>
      <item path="" reason=""></item>
    </how_to_read>
    <entry_points>
      <entry path="" symbol="" kind=""></entry>
    </entry_points>
  </reading_guide>
  <artifacts>
    <documents>
      <document path="" kind=""></document>
    </documents>
    <code>
      <file path="">
        <symbol name="" kind="" />
      </file>
    </code>
    <logs>
      <log path=""></log>
    </logs>
    <inputs>
      <input path="" kind=""></input>
    </inputs>
  </artifacts>
  <observations>
    <fact></fact>
  </observations>
  <handoff>
    <next_skill></next_skill>
    <must_read>
      <path></path>
    </must_read>
  </handoff>
</context_packet>
```

## Rules

- `scope` には、指示された走査範囲、実際に走査した範囲、未走査範囲だけを書く。
- `reading_guide` には、次の skill がどこから読めばよいかをパスと symbol 名だけで書く。
- `artifacts` には、存在確認できた docs、code、logs、inputs を列挙する。
- `observations` には、コード・仕様・ログ・入力から直接確認できた事実だけを書く。
- `handoff` には、次の skill 名と、その skill が最低限読むべきパスだけを書く。
- 値が無い節は空要素ではなく節ごと省略してよい。
- 長文引用は禁止する。
- 1 回読めば次の skill が着手できる密度に圧縮する。
