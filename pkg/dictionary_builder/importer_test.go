package dictionary_builder_test

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary_builder"
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

	// Create table
	query := `
	CREATE TABLE IF NOT EXISTS dictionary_entries (
		edid TEXT PRIMARY KEY,
		rec TEXT NOT NULL,
		source TEXT NOT NULL,
		dest TEXT NOT NULL,
		addon TEXT NOT NULL
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	store := dictionary_builder.NewDictionaryStore(db)
	config := dictionary_builder.DefaultConfig()
	importer := dictionary_builder.NewImporter(config, store, slog.Default())

	ctx := context.Background()
	// Use strings.NewReader directly instead of a temp file
	file := strings.NewReader(dummyXML)

	count, err := importer.ImportXML(ctx, file)

	require.NoError(t, err)
	assert.Equal(t, 3, count, "Should have imported exactly 3 valid records (resolving 1 duplicate UPSERT)")

	rows, err := db.Query("SELECT edid, rec, source, dest, addon FROM dictionary_entries ORDER BY edid")
	require.NoError(t, err)
	defer rows.Close()

	var entries []dictionary_builder.DictTerm
	for rows.Next() {
		var e dictionary_builder.DictTerm
		err := rows.Scan(&e.EDID, &e.REC, &e.Source, &e.Dest, &e.Addon)
		require.NoError(t, err)
		entries = append(entries, e)
	}

	assert.Len(t, entries, 2)

	assert.Equal(t, "Skyrim.esm|0x0001", entries[0].EDID)
	assert.Equal(t, "BOOK:FULL", entries[0].REC)
	assert.Equal(t, "アルゴニアンの侍女 v2", entries[0].Dest)
	assert.Equal(t, "Skyrim.esm", entries[0].Addon)

	assert.Equal(t, "Skyrim.esm|0x0002", entries[1].EDID)
	assert.Equal(t, "NPC_:FULL", entries[1].REC)
	assert.Equal(t, "ウルフリック・ストームクローク", entries[1].Dest)
	assert.Equal(t, "Skyrim.esm", entries[1].Addon)
}
