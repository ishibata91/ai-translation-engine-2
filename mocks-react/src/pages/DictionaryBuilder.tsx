import React from 'react';

const DictionaryBuilder: React.FC = () => {
    return (
        <div className="flex flex-col w-full p-4 gap-4">
            {/* ヘッダー部分 */}
            <div className="navbar bg-base-100 rounded-box shadow-sm px-4">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">辞書構築 (Dictionary Builder)</span>
                </div>
            </div>

            {/* 画面説明 (通知エリア) */}
            <div className="alert alert-info shadow-sm flex-col items-start gap-2">
                <div className="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                    <h3 className="font-bold">システム辞書の構築について</h3>
                </div>
                <div className="text-sm space-y-2">
                    <p>
                        この画面では、公式翻訳や過去の翻訳済みModのデータ（SSTXML形式など）をインポートし、<strong>全プロジェクト共通で利用される「システム辞書(dictionary.db)」</strong>を構築・管理します。
                    </p>
                    <ul className="list-disc list-inside ml-2">
                        <li>システム辞書に登録された原文と訳文のペアは、AI翻訳時に抽出され、<strong>公式用語や固有名詞の訳揺れを防ぐための重要なコンテキスト</strong>として活用されます。</li>
                        <li><code className="bg-base-200 px-1 rounded">Skyrim.esm</code> などの公式マスターファイルや、大規模ModのSSTXMLを優先してインポートすることを強く推奨します。</li>
                        <li>現在、辞書データベースの準備は整っています。「XMLインポート」から新しいファイルを指定してインポートを開始してください。</li>
                    </ul>
                </div>
            </div>

            {/* 上部パネル */}
            <div className="grid grid-cols-2 gap-4">
                {/* インポート設定カード */}
                <div className="card bg-base-100 shadow-sm border border-base-200">
                    <div className="card-body">
                        <h2 className="card-title text-base">XMLインポート (xTranslator形式)</h2>
                        <div className="flex flex-col gap-4 mt-2">
                            <span className="text-sm">SSTXMLファイル、または公式DLCの翻訳XMLを選択してください。選択後、自動的にインポートが開始されます。</span>
                            <div className="flex gap-4">
                                <input type="file" className="file-input file-input-bordered file-input-primary w-full max-w-xs" />
                            </div>
                            <div>
                                <span className="text-sm font-bold block mb-2">インポート進捗</span>
                                <progress className="progress progress-primary w-full" value="0" max="100"></progress>
                            </div>
                        </div>
                    </div>
                </div>

                {/* 辞書統計カード */}
                <div className="card bg-base-100 shadow-sm border border-base-200">
                    <div className="card-body">
                        <h2 className="card-title text-base">システム辞書ステータス</h2>
                        <div className="flex flex-col gap-4 mt-2">
                            <div className="stat px-0 py-2">
                                <div className="stat-title text-sm">総エントリ数</div>
                                <div className="stat-value text-3xl font-mono">1,240,512</div>
                            </div>
                            <div className="stat px-0 py-2 border-t border-base-200">
                                <div className="stat-title text-sm">登録済みソース</div>
                                <div className="stat-value text-xl">5</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* ボトムパネル: 登録済みソース一覧 */}
            <div className="card w-full bg-base-100 shadow-sm border border-base-200 flex-1">
                <div className="card-body">
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="card-title text-base">登録済み辞書ソース一覧</h2>
                        <button className="btn btn-outline btn-error btn-sm">全て削除</button>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="table table-zebra w-full">
                            <thead>
                                <tr>
                                    <th>ソース名 (ファイル名)</th>
                                    <th>エントリ数</th>
                                    <th>最終更新日時</th>
                                    <th>ステータス</th>
                                    <th>アクション</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>Skyrim.esm (SSTXML)</td>
                                    <td className="font-mono text-right">850,231</td>
                                    <td>2026-02-26 12:00</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs text-error">削除</button></td>
                                </tr>
                                <tr>
                                    <td>Update.esm (SSTXML)</td>
                                    <td className="font-mono text-right">10,023</td>
                                    <td>2026-02-26 12:01</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs text-error">削除</button></td>
                                </tr>
                                <tr>
                                    <td>Dawnguard.esm (SSTXML)</td>
                                    <td className="font-mono text-right">150,490</td>
                                    <td>2026-02-26 12:05</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs text-error">削除</button></td>
                                </tr>
                                <tr>
                                    <td>HearthFires.esm (SSTXML)</td>
                                    <td className="font-mono text-right">25,102</td>
                                    <td>2026-02-26 12:06</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs text-error">削除</button></td>
                                </tr>
                                <tr>
                                    <td>Dragonborn.esm (SSTXML)</td>
                                    <td className="font-mono text-right">204,666</td>
                                    <td>2026-02-26 12:10</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs text-error">削除</button></td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DictionaryBuilder;
