import React, { useState } from 'react';

interface DictSourceRow {
    id: string;
    fileName: string;
}

interface CrossSearchModalProps {
    sources: DictSourceRow[];
    onSearch: (query: string) => void;
    onClose: () => void;
}

/**
 * CrossSearchModal: 全辞書ソースを横断して検索するためのモーダル。
 * キーワードを入力して「検索実行」を押すと DictionaryBuilder に query を返す。
 */
const CrossSearchModal: React.FC<CrossSearchModalProps> = ({ sources, onSearch, onClose }) => {
    const [query, setQuery] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!query.trim()) return;
        onSearch(query.trim());
    };

    return (
        <dialog open className="modal modal-open">
            <div className="modal-box max-w-lg">
                <h3 className="font-bold text-lg mb-4">🔎 辞書横断検索</h3>
                <p className="text-sm text-base-content/70 mb-4">
                    登録されている全辞書ソース ({sources.length} 件) を対象に、
                    原文・訳文・Editor ID を横断して検索します。
                </p>

                <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                    <div className="form-control">
                        <label className="label">
                            <span className="label-text font-bold">検索キーワード</span>
                        </label>
                        <input
                            type="text"
                            className="input input-bordered w-full"
                            placeholder="例: 魔法、dragon、FireBall ..."
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            autoFocus
                        />
                        <label className="label">
                            <span className="label-text-alt text-base-content/50">
                                部分一致で検索します。最大 500 件/ページで取得します。
                            </span>
                        </label>
                    </div>

                    {sources.length > 0 && (
                        <div className="rounded-lg border border-base-200 bg-base-200/40 p-3">
                            <p className="text-xs font-bold text-base-content/60 mb-2">検索対象の辞書ソース</p>
                            <div className="flex flex-wrap gap-1 max-h-28 overflow-y-auto">
                                {sources.map(s => (
                                    <span key={s.id} className="badge badge-ghost badge-sm font-mono">{s.fileName}</span>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="modal-action mt-0">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>キャンセル</button>
                        <button
                            type="submit"
                            className={`btn btn-primary ${!query.trim() ? 'btn-disabled' : ''}`}
                            disabled={!query.trim()}
                        >
                            🔎 検索実行
                        </button>
                    </div>
                </form>
            </div>
            <div className="modal-backdrop" onClick={onClose} />
        </dialog>
    );
};

export default CrossSearchModal;
