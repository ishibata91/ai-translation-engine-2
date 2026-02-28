import React from 'react';

const MasterPersona: React.FC = () => {
    const showModal = (id: string) => {
        const modal = document.getElementById(id) as HTMLDialogElement;
        modal?.showModal();
    };

    return (
        <div className="flex flex-col w-full p-4 gap-4 pb-20">
            {/* ヘッダー部分 */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">マスターペルソナ構築 (Master Persona Builder)</span>
                </div>
            </div>

            {/* 通知エリア */}
            <div className="alert alert-info shadow-sm">
                <span><code>extractData.pas</code>
                    で抽出されたベースゲームのJSONデータからNPCのセリフを解析し、LLMを用いて基本となるペルソナ（性格・口調）を生成・キャッシュします。これによりMod翻訳時の品質と一貫性が向上します。</span>
            </div>

            {/* 上部パネル */}
            <div className="grid grid-cols-2 gap-4">
                {/* 生成設定カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">JSONデータのインポートと生成</h2>
                        <div className="flex flex-col gap-4 mt-2">
                            <span className="text-sm">xEditスクリプト <code>extractData.pas</code>
                                によって抽出された、マスターファイルのJSONデータを選択し、ペルソナ生成を開始します。</span>
                            <div className="flex gap-4 items-center">
                                <input type="file" accept=".json"
                                    className="file-input file-input-bordered file-input-primary w-full max-w-xs" />
                                <button className="btn btn-primary">アップロード</button>
                            </div>
                            <div>
                                <span className="mt-2 mb-1 block text-sm text-base-content/70 font-bold">全体進捗</span>
                                <progress className="progress progress-primary w-full" value="45" max="100"></progress>
                            </div>
                        </div>
                    </div>
                </div>

                {/* 統計カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">ペルソナDB ステータス</h2>
                        <div className="grid grid-cols-2 gap-4 mt-2">
                            <div className="stat p-0">
                                <div className="stat-title text-sm">登録済みNPC数</div>
                                <div className="stat-value text-primary font-mono text-3xl">2,451</div>
                            </div>
                            <div className="stat p-0">
                                <div className="stat-title text-sm">生成エラー</div>
                                <div className="stat-value text-error font-mono text-3xl">12</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* ボトムパネル: NPC処理リスト */}
            <div className="card w-full bg-base-100 border border-base-200 shadow-sm flex-1">
                <div className="card-body flex flex-col overflow-hidden">
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="card-title text-base">NPC処理ステータス (Skyrim.esm)</h2>
                        <div className="flex gap-2">
                            <input type="text" placeholder="FormID / Name..." className="input input-bordered input-sm" />
                            <select className="select select-bordered select-sm">
                                <option>すべて</option>
                                <option>抽出完了</option>
                                <option>生成中</option>
                                <option>完了</option>
                                <option>エラー</option>
                            </select>
                        </div>
                    </div>
                    <div className="overflow-x-auto flex-1">
                        <table className="table table-pin-rows table-zebra w-full">
                            <thead>
                                <tr>
                                    <th>FormID</th>
                                    <th>NPC名 (EditorID)</th>
                                    <th>セリフ数</th>
                                    <th>ステータス</th>
                                    <th>生成日時</th>
                                    <th>詳細・アクション</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td className="font-mono text-sm">00013B9B</td>
                                    <td>UlfricStormcloak</td>
                                    <td className="font-mono text-right">342</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td>2026-02-26 14:02</td>
                                    <td>
                                        <div className="flex gap-2">
                                            <button className="btn btn-ghost btn-xs" onClick={() => showModal('persona_modal')}>ペルソナ確認</button>
                                            <button className="btn btn-ghost btn-xs text-error" onClick={() => showModal('delete_modal')}>削除</button>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="font-mono text-sm">0001A694</td>
                                    <td>Tullius</td>
                                    <td className="font-mono text-right">298</td>
                                    <td>
                                        <div className="badge badge-success badge-sm">完了</div>
                                    </td>
                                    <td>2026-02-26 14:05</td>
                                    <td>
                                        <div className="flex gap-2">
                                            <button className="btn btn-ghost btn-xs" onClick={() => showModal('persona_modal')}>ペルソナ確認</button>
                                            <button className="btn btn-ghost btn-xs text-error" onClick={() => showModal('delete_modal')}>削除</button>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="font-mono text-sm">0001A695</td>
                                    <td>Rikke</td>
                                    <td className="font-mono text-right">156</td>
                                    <td>
                                        <div className="badge badge-info badge-sm">生成中</div>
                                    </td>
                                    <td>-</td>
                                    <td>
                                        <div className="flex gap-2">
                                            <button className="btn btn-ghost btn-xs" disabled>処理中...</button>
                                            <button className="btn btn-ghost btn-xs text-error" disabled>削除</button>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="font-mono text-sm">00013BA1</td>
                                    <td>BalgruufTheGreater</td>
                                    <td className="font-mono text-right">210</td>
                                    <td>
                                        <div className="badge badge-ghost badge-sm">抽出完了</div>
                                    </td>
                                    <td>-</td>
                                    <td>
                                        <div className="flex gap-2">
                                            <button className="btn btn-primary btn-xs" onClick={() => showModal('generate_modal')}>生成開始</button>
                                            <button className="btn btn-ghost btn-xs text-error" onClick={() => showModal('delete_modal')}>削除</button>
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="font-mono text-sm">0001A696</td>
                                    <td>GalmarStoneFist</td>
                                    <td className="font-mono text-right">124</td>
                                    <td>
                                        <div className="badge badge-error badge-sm">エラー</div>
                                    </td>
                                    <td>2026-02-26 14:10</td>
                                    <td>
                                        <div className="flex gap-2">
                                            <button className="btn btn-ghost btn-xs font-bold text-error">再試行</button>
                                            <button className="btn btn-ghost btn-xs text-error" onClick={() => showModal('delete_modal')}>削除</button>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* ペルソナ確認モーダル */}
            <dialog id="persona_modal" className="modal">
                <div className="modal-box w-11/12 max-w-3xl border border-base-300">
                    <h3 className="font-bold text-lg">ペルソナ詳細確認: UlfricStormcloak (00013B9B)</h3>
                    <div className="py-4 flex flex-col gap-4">
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">要約 (Summary)</span></label>
                            <textarea className="textarea textarea-bordered h-24"
                                readOnly value="ストームクロークの反乱軍のリーダー。誇り高く、ノルドの伝統を重んじる。ウィンドヘルムの首長であり、帝国に強い敵対心を抱いている。"></textarea>
                        </div>
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">口調・一人称・二人称 (Tone/Pronouns)</span></label>
                            <textarea className="textarea textarea-bordered h-24" readOnly value={`一人称：「俺」\n二人称：「お前」「お前たち」\n口調：威厳があり、力強く、少し乱暴な言葉遣い。命令形をよく使う。`}></textarea>
                        </div>
                    </div>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-outline">再生成</button>
                                <button className="btn btn-primary">閉じる</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {/* 生成モーダル */}
            <dialog id="generate_modal" className="modal">
                <div className="modal-box w-11/12 max-w-4xl border border-base-300">
                    <h3 className="font-bold text-lg">ペルソナ生成: BalgruufTheGreater (00013BA1)</h3>
                    <div className="py-4 flex flex-col gap-4 h-[60vh]">
                        <div className="alert alert-info shadow-sm py-2">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                className="stroke-info shrink-0 w-6 h-6">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                            </svg>
                            <div>
                                <h3 className="font-bold text-sm">予測コンテキスト (Estimated Tokens)</h3>
                                <div className="text-xs">入力トークン: 約 3,500 Tokens / 会話行数: 210行</div>
                            </div>
                        </div>

                        <div className="form-control flex-1 overflow-hidden">
                            <label className="label py-1"><span className="label-text font-bold">抽出されたセリフ (会話文)</span></label>
                            <div className="mockup-code h-full overflow-y-auto bg-base-200 text-base-content text-sm relative">
                                <pre data-prefix=">"><code>"ああ、ドラゴンボーン。よく来てくれたな"</code></pre>
                                <pre data-prefix=">"><code>"ホワイトランの民は、お前の働きに深く感謝しているぞ"</code></pre>
                                <pre data-prefix=">"><code>"首長としての私の立場は複雑だ。バルグルーフはどちらの味方なのか、とよく聞かれる"</code></pre>
                                <pre data-prefix=">"><code>"だが私の使命は、何よりもまずホワイトランを守ることにある"</code></pre>
                                <pre data-prefix=">"><code>"イリレス、衛兵の巡回を強化しろ。"</code></pre>
                                <pre data-prefix=">"><code>"......"</code></pre>
                            </div>
                        </div>
                    </div>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-primary">この内容で生成を実行する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {/* 削除確認モーダル */}
            <dialog id="delete_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">削除の確認</h3>
                    <p className="py-4">このNPCデータを一覧およびデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-error">削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>
        </div>
    );
};

export default MasterPersona;
