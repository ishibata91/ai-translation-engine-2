package dictionary_test

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dummyXML = `<?xml version="1.0" encoding="utf-8"?>
<SSTXMLRessources>
  <Params>
    <Addon>Skyrim.esm</Addon>
    <Source>english</Source>
    <Dest>japanese</Dest>
  </Params>
  <Content>
    <String>
      <EDID>Skyrim.esm|0x0001</EDID>
      <REC>BOOK:FULL</REC>
      <Source>The Lusty Argonian Maid</Source>
      <Dest>アルゴニアンの侍女</Dest>
    </String>
    <String>
      <EDID>Skyrim.esm|0x0002</EDID>
      <REC>NPC_:FULL</REC>
      <Source>Ulfric Stormcloak</Source>
      <Dest>ウルフリック・ストームクローク</Dest>
    </String>
    <String>
      <EDID>Skyrim.esm|0x0003</EDID>
      <REC>INFO</REC>
      <Source>I used to be an adventurer like you.</Source>
      <Dest>昔はお前のような冒険者だったのだが。</Dest>
    </String>
    <String>
      <EDID>Skyrim.esm|0x0001</EDID>
      <REC>BOOK:FULL</REC>
      <Source>The Lusty Argonian Maid</Source>
      <Dest>アルゴニアンの侍女 v2</Dest>
    </String>
  </Content>
</SSTXMLRessources>
`

func TestImporter_ImportXML(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize new schema
	store, err := dictionary.NewDictionaryStore(db)
	require.NoError(t, err)

	config := dictionary.DefaultConfig()
	notifier := progress.NewNoopNotifier()
	importer := dictionary.NewImporter(config, store, notifier, slog.Default())

	ctx := context.Background()
	// Use strings.NewReader directly instead of a temp file
	file := strings.NewReader(dummyXML)

	// Create a dummy source first
	src := &dictionary.DictSource{
		FileName: "Skyrim.esm",
		Format:   "xml",
		FilePath: "dummy.xml",
		FileSize: int64(len(dummyXML)),
		Status:   "PENDING",
	}
	sourceID, err := store.CreateSource(ctx, src)
	require.NoError(t, err)

	count, err := importer.ImportXML(ctx, sourceID, "dummy.xml", file)

	require.NoError(t, err)
	assert.Equal(t, 3, count, "Should have imported exactly 3 valid strings (no UPSERT logic anymore)")

	// In the new schema without UPSERT logic in SaveTerms, it will just insert 3 rows if there are 3 valid strings.
	// In the dummy data:
	// 0x0001 (BOOK:FULL)
	// 0x0002 (NPC_:FULL)
	// 0x0003 (INFO) -> Not in default config
	// 0x0001 (BOOK:FULL)
	// So 3 records should be inserted in total.
	rows, err := db.Query("SELECT edid, record_type, source_text, dest_text FROM dlc_dictionary_entries ORDER BY edid, dest_text")
	require.NoError(t, err)
	defer rows.Close()

	var entries []dictionary.DictTerm
	for rows.Next() {
		var e dictionary.DictTerm
		err := rows.Scan(&e.EDID, &e.RecordType, &e.Source, &e.Dest)
		require.NoError(t, err)
		entries = append(entries, e)
	}

	assert.Len(t, entries, 3)

	assert.Equal(t, "Skyrim.esm|0x0001", entries[0].EDID)
	assert.Equal(t, "BOOK:FULL", entries[0].RecordType)
	assert.Equal(t, "アルゴニアンの侍女", entries[0].Dest)

	assert.Equal(t, "Skyrim.esm|0x0001", entries[1].EDID)
	assert.Equal(t, "BOOK:FULL", entries[1].RecordType)
	assert.Equal(t, "アルゴニアンの侍女 v2", entries[1].Dest)

	assert.Equal(t, "Skyrim.esm|0x0002", entries[2].EDID)
	assert.Equal(t, "NPC_:FULL", entries[2].RecordType)
	assert.Equal(t, "ウルフリック・ストームクローク", entries[2].Dest)
}
