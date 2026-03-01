package dictionary

import "time"

// DictSource は辞書ソースファイルのメタデータを表す。
// dlc_sources テーブルに対応する。
type DictSource struct {
	ID           int64      `json:"id"`
	FileName     string     `json:"file_name"`
	Format       string     `json:"format"`
	FilePath     string     `json:"file_path"`
	FileSize     int64      `json:"file_size_bytes"`
	EntryCount   int        `json:"entry_count"`
	Status       string     `json:"status"` // PENDING, IMPORTING, COMPLETED, ERROR
	ErrorMessage string     `json:"error_message,omitempty"`
	ImportedAt   *time.Time `json:"imported_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// DictTerm は辞書の個別単語エントリを表す。
// dlc_dictionary_entries テーブルに対応する。
type DictTerm struct {
	ID         int64  `json:"id"`
	SourceID   int64  `json:"source_id"`
	EDID       string `json:"edid"`
	RecordType string `json:"record_type"`
	Source     string `json:"source_text"`
	Dest       string `json:"dest_text"`
}
