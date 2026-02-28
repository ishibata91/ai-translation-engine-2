import React, { useState } from 'react';

const TranslationFlow: React.FC = () => {
    const [activeTab, setActiveTab] = useState(0);

    const tabs = [
        { label: '用語' },
        { label: 'ペルソナ生成' },
        { label: '要約' },
        { label: '翻訳' },
        { label: 'エクスポート' }
    ];

    return (
        <div className="flex flex-col w-full p-4 gap-4 h-full">
            {/* ヘッダー部分 */}
            <div className="flex justify-between items-center w-full bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <div className="flex items-center gap-4">
                    <span className="text-2xl font-bold">翻訳プロジェクト: Skyrim.esm</span>
                    <span className="badge badge-primary badge-lg badge-outline">進行中</span>
                </div>
                <button className="btn btn-outline btn-sm">プロジェクト設定</button>
            </div>

            {/* ステッパー (全体の進捗) */}
            <div className="bg-base-100 p-4 rounded-xl shadow-sm border border-base-200 shrink-0">
                <ul className="steps w-full">
                    {tabs.map((tab, idx) => (
                        <li key={idx} className={`step ${idx <= activeTab ? 'step-primary' : ''}`}>{tab.label}</li>
                    ))}
                </ul>
            </div>

            <div className="bg-base-100 rounded-xl shadow-sm border border-base-200 flex flex-col flex-1 overflow-hidden">
                {/* タブナビゲーション */}
                <div role="tablist" className="tabs tabs-bordered w-full pt-2 shrink-0">
                    {tabs.map((tab, idx) => (
                        <button
                            key={idx}
                            role="tab"
                            className={`tab ${activeTab === idx ? 'tab-active' : ''}`}
                            onClick={() => setActiveTab(idx)}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                {/* ▼▼▼ タブ・コンテンツ群 ▼▼▼ */}
                <div className="flex flex-col p-4 flex-1 overflow-hidden relative">

                    {/* 1. 用語パネル (Terminology) */}
                    <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${activeTab === 0 ? 'flex' : 'hidden'}`}>
                        <div className="alert alert-info shadow-sm shrink-0">
                            <span>Mod内から抽出された固有名詞や特殊用語を翻訳し、Mod専用辞書を構築・保存しています。</span>
                            <div className="flex-none">
                                <button className="btn btn-sm btn-primary" onClick={() => setActiveTab(1)}>用語翻訳を開始</button>
                            </div>
                        </div>

                        {/* 統計サマリー */}
                        <div className="flex gap-4 shrink-0">
                            <div className="stat bg-base-100 border rounded-xl shadow-sm p-4">
                                <div className="stat-title text-xs font-bold">抽出用語数</div>
                                <div className="stat-value text-2xl">1,245</div>
                                <div className="stat-desc">一意な抽出レコード</div>
                            </div>
                            <div className="stat bg-base-100 border rounded-xl shadow-sm p-4">
                                <div className="stat-title text-xs font-bold">辞書マッチ (強制翻訳)</div>
                                <div className="stat-value text-success text-2xl">892</div>
                                <div className="stat-desc">LLMを介さず確定</div>
                            </div>
                            <div className="stat bg-base-100 border rounded-xl shadow-sm p-4">
                                <div className="stat-title text-xs font-bold">LLM翻訳</div>
                                <div className="stat-value text-primary text-2xl">353</div>
                                <div className="stat-desc">新規・未知の用語</div>
                            </div>
                            <div className="stat bg-base-100 border rounded-xl shadow-sm p-4">
                                <div className="stat-figure text-primary">
                                    <div className="radial-progress text-sm font-bold" style={{ '--value': 80, '--size': '3rem' } as React.CSSProperties}>80%</div>
                                </div>
                                <div className="stat-title text-xs font-bold">全体進捗</div>
                                <div className="stat-value text-2xl">996</div>
                                <div className="stat-desc">完了した用語</div>
                            </div>
                        </div>

                        {/* 翻訳結果リスト */}
                        <div className="overflow-y-auto border rounded-xl flex-1 bg-base-100">
                            <table className="table table-zebra table-pin-rows w-full">
                                <thead>
                                    <tr className="bg-base-200">
                                        <th className="w-1/6">種別</th>
                                        <th className="w-2/6">翻訳対象 (Source)</th>
                                        <th className="w-12 text-center"></th>
                                        <th className="w-2/6">結果 (Result)</th>
                                        <th className="w-1/6">状態 / ロジック</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td><span className="badge badge-outline badge-sm text-xs font-mono">LCTN:FULL</span></td>
                                        <td className="font-mono text-sm">Whiterun</td>
                                        <td className="text-center text-gray-400">➔</td>
                                        <td className="font-bold text-sm">ホワイトラン</td>
                                        <td><span className="badge badge-success badge-sm">辞書・完全一致</span></td>
                                    </tr>
                                    <tr>
                                        <td><span className="badge badge-outline badge-sm text-xs font-mono">NPC_:FULL</span></td>
                                        <td className="font-mono text-sm">Jon Battle-Born</td>
                                        <td className="text-center text-gray-400">➔</td>
                                        <td className="font-bold text-sm">ジョン・バトルボーン</td>
                                        <td><span className="badge badge-success badge-sm">辞書・部分一致</span></td>
                                    </tr>
                                    <tr>
                                        <td><span className="badge badge-outline badge-sm text-xs font-mono">WEAP:FULL</span></td>
                                        <td className="font-mono text-sm">Skyforge Steel Sword</td>
                                        <td className="text-center text-gray-400">➔</td>
                                        <td className="font-bold text-sm">スカイフォージの鋼鉄の剣</td>
                                        <td><span className="badge badge-primary badge-sm text-white">LLM翻訳・完了</span></td>
                                    </tr>
                                    <tr>
                                        <td><span className="badge badge-outline badge-sm text-xs font-mono">NPC_:SHRT</span></td>
                                        <td className="font-mono text-sm">Ulfric</td>
                                        <td className="text-center text-gray-400">➔</td>
                                        <td className="text-gray-400 italic text-sm">翻訳中...</td>
                                        <td><span className="badge badge-ghost badge-sm border-base-300">LLM取得中...</span></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                            <span className="text-sm font-bold text-gray-500 ml-2">Job: TerminologyTranslation (Running)</span>
                            <div className="flex gap-2">
                                <button className="btn btn-ghost btn-sm">一時停止</button>
                                <button className="btn btn-primary btn-sm" onClick={() => setActiveTab(1)}>用語を確定して次へ</button>
                            </div>
                        </div>
                    </div>

                    {/* 2. ペルソナ生成パネル (Persona) */}
                    <div className={`tab-content-panel flex-col gap-4 h-full ${activeTab === 1 ? 'flex' : 'hidden'}`}>
                        <div className="alert alert-info shadow-sm shrink-0">
                            <span>検出されたNPCのペルソナ（性格・口調）を生成します。マスター辞書に存在すればキャッシュを利用します。</span>
                            <div className="flex-none">
                                <button className="btn btn-sm btn-primary">LLMで一括生成</button>
                            </div>
                        </div>
                        <div className="flex gap-4 flex-1 min-h-0 overflow-hidden">
                            {/* 左：NPCリスト */}
                            <div className="w-1/3 border rounded-xl bg-base-100 overflow-y-auto">
                                <ul className="menu w-full bg-base-100">
                                    <li className="menu-title">NPC一覧 (24)</li>
                                    <li><a className="active">Ulfric Stormcloak <span className="badge badge-success badge-sm ml-auto">完了</span></a></li>
                                    <li><a>General Tullius <span className="badge badge-success badge-sm ml-auto">完了</span></a></li>
                                    <li><a>Elisif the Fair <span className="badge badge-warning badge-sm ml-auto">生成中</span></a></li>
                                    <li><a>Whiterun Guard <span className="badge badge-ghost badge-sm ml-auto">未生成</span></a></li>
                                </ul>
                            </div>
                            {/* 右：ペルソナ詳細 */}
                            <div className="w-2/3 border rounded-xl bg-base-100 p-4 flex flex-col gap-4 overflow-y-auto">
                                <h3 className="text-xl font-bold border-b pb-2">Ulfric Stormcloak (0001414D)</h3>
                                <div className="flex gap-2">
                                    <span className="badge badge-outline">Race: Nord</span>
                                    <span className="badge badge-outline">Sex: Male</span>
                                    <span className="badge badge-outline">Class: Warrior</span>
                                </div>
                                <div className="form-control flex-1 flex flex-col min-h-0">
                                    <label className="label"><span className="label-text font-bold">生成されたペルソナ情報 (プロンプト動的注入用)</span></label>
                                    <textarea className="textarea textarea-bordered flex-1 text-base leading-relaxed" defaultValue={`誇り高く、カリスマ性のあるストームクロークの反乱軍リーダー。雄弁で威厳のある口調。\n一人称: 私、俺\n二人称: お前、貴様\n特徴: スカイリムの独立とノルドの誇りを強調する。「～だ」「～だろう」「～なのだ」といった断定的で力強い語尾を多用する。若者には厳しくも期待を込めて接する。`}></textarea>
                                </div>
                                <div className="flex justify-end gap-2 mt-2">
                                    <button className="btn btn-outline btn-sm">再生成</button>
                                    <button className="btn btn-secondary btn-sm">保存</button>
                                </div>
                            </div>
                        </div>
                        <div className="flex justify-end gap-2 shrink-0">
                            <button className="btn btn-primary" onClick={() => setActiveTab(2)}>ペルソナを確定して次へ</button>
                        </div>
                    </div>

                    {/* 3. 要約パネル (Summary) */}
                    <div className={`tab-content-panel flex-col gap-4 h-full ${activeTab === 2 ? 'flex' : 'hidden'}`}>
                        <div className="alert alert-info shadow-sm shrink-0">
                            <span>長文の書籍や連続する会話シーンの要約を生成し、翻訳時のコンテキストとして利用します。</span>
                            <div className="flex-none">
                                <button className="btn btn-sm btn-primary">要約の生成開始</button>
                            </div>
                        </div>
                        <div className="overflow-y-auto border rounded-xl flex-1 bg-base-100">
                            <table className="table table-zebra table-pin-rows w-full">
                                <thead>
                                    <tr>
                                        <th>種別</th>
                                        <th>対象レコード/シーン</th>
                                        <th>状態</th>
                                        <th>要約内容</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td><span className="badge badge-primary badge-sm text-white">Book</span></td>
                                        <td>The Lusty Argonian Maid, v1</td>
                                        <td><span className="badge badge-success badge-sm">完了</span></td>
                                        <td className="text-sm">アルゴニアンのメイドであるリフトスと、主人のクラシウス・キュリオの際どい会話劇。比喩表現が多用される。</td>
                                    </tr>
                                    <tr>
                                        <td><span className="badge badge-secondary badge-sm text-white">Dialog</span></td>
                                        <td>MQ101_UlfricExecution</td>
                                        <td><span className="badge badge-success badge-sm">完了</span></td>
                                        <td className="text-sm">ヘルゲンでの処刑シーン。帝国軍によるストームクローク兵の処刑と、突然のドラゴンの襲撃。緊迫した雰囲気。</td>
                                    </tr>
                                    <tr>
                                        <td><span className="badge badge-secondary badge-sm text-white">Dialog</span></td>
                                        <td>MQ102_RiverwoodArrive</td>
                                        <td><span className="badge badge-ghost badge-sm">未生成</span></td>
                                        <td className="text-gray-400 italic">（未生成）</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        <div className="flex justify-end gap-2 shrink-0">
                            <button className="btn btn-primary" onClick={() => setActiveTab(3)}>要約を確定して次へ</button>
                        </div>
                    </div>

                    {/* 4. 翻訳パネル (Translation) */}
                    <div className={`tab-content-panel gap-4 h-full min-h-0 overflow-hidden ${activeTab === 3 ? 'flex' : 'hidden'}`}>
                        {/* 左ペイン: 翻訳対象リスト */}
                        <div className="w-1/2 flex flex-col border rounded-xl bg-base-100 overflow-hidden">
                            <div className="p-3 border-b flex flex-col gap-2 bg-base-200 shrink-0">
                                <div className="text-sm font-bold">レコード一覧 (全 1,200 件)</div>
                                <div className="flex gap-2">
                                    <input type="text" placeholder="原文・訳文を検索..." className="input input-bordered w-full input-sm" />
                                    <select className="select select-bordered select-sm">
                                        <option>未翻訳 (800)</option>
                                        <option>要確認 (15)</option>
                                        <option>すべて表示</option>
                                    </select>
                                </div>
                            </div>

                            <div className="overflow-y-auto flex-1 h-full">
                                <table className="table table-zebra table-pin-rows w-full select-none table-sm">
                                    <thead>
                                        <tr>
                                            <th>EDID</th>
                                            <th>原文</th>
                                            <th>状態</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr className="hover cursor-pointer bg-base-300">
                                            <td>0001A2B4</td>
                                            <td className="truncate max-w-[200px]">Then I took an arrow in the knee.</td>
                                            <td><span className="badge badge-warning badge-xs">要確認</span></td>
                                        </tr>
                                        <tr className="hover cursor-pointer">
                                            <td>0001A2B3</td>
                                            <td className="truncate max-w-[200px]">I used to be an adventurer like you.</td>
                                            <td><span className="badge badge-success badge-xs">完了</span></td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                            <div className="p-2 border-t flex justify-center bg-base-200 shrink-0">
                                <div className="join">
                                    <button className="join-item btn btn-xs">«</button>
                                    <button className="join-item btn btn-xs">Page 1</button>
                                    <button className="join-item btn btn-xs">»</button>
                                </div>
                            </div>
                        </div>

                        {/* 右ペイン: 翻訳コンソール (詳細) */}
                        <div className="w-1/2 flex flex-col rounded-xl border bg-base-100 relative overflow-hidden">
                            {/* ヘッダーアクション */}
                            <div className="shrink-0 z-10 bg-base-100 border-b p-3 flex justify-between items-center shadow-sm">
                                <span className="font-bold">自動翻訳コンソール</span>
                                <div className="flex gap-2">
                                    <button className="btn btn-outline btn-xs">自動進行を開始</button>
                                </div>
                            </div>

                            <div className="p-4 flex flex-col gap-4 flex-1 overflow-y-auto">
                                <div className="flex justify-between items-center text-sm shrink-0">
                                    <span className="badge badge-outline">Record: 0001A2B4</span>
                                </div>

                                {/* コンテキスト情報 */}
                                <div className="collapse collapse-arrow bg-base-200 border border-base-300 rounded-box shrink-0">
                                    <input type="checkbox" defaultChecked />
                                    <div className="collapse-title font-bold text-sm">
                                        適用されたコンテキスト (Context DTO)
                                    </div>
                                    <div className="collapse-content flex flex-col gap-2">
                                        <div className="alert shadow-sm p-2 flex-row justify-start gap-4 rounded-lg bg-base-100 border border-base-200 text-left">
                                            <span className="badge badge-primary badge-sm text-white shrink-0">要約</span>
                                            <span className="text-xs">衛兵が過去の冒険譚を語るシーン。</span>
                                        </div>
                                        <div className="alert shadow-sm p-2 flex-row justify-start gap-4 rounded-lg bg-base-100 border border-base-200 text-left">
                                            <span className="badge badge-secondary badge-sm text-white shrink-0">ペルソナ</span>
                                            <div className="flex flex-col">
                                                <span className="text-xs font-bold">MaleGuard</span>
                                                <span className="text-xs">ぶっきらぼうだが親しみやすい口調。「～だ」</span>
                                            </div>
                                        </div>
                                        <div className="alert shadow-sm p-2 flex-row justify-start gap-4 rounded-lg bg-base-100 border border-base-200 text-left">
                                            <span className="badge badge-accent badge-sm text-white shrink-0">用語</span>
                                            <span className="text-xs font-mono">arrow in the knee</span> <span className="text-xs">➡️ 膝に矢を受けて</span>
                                        </div>
                                    </div>
                                </div>

                                {/* 原文と翻訳 */}
                                <div className="flex flex-col gap-1 shrink-0">
                                    <label className="label pb-0"><span className="label-text font-bold">原文 (Source)</span></label>
                                    <div className="p-3 bg-base-200 rounded-lg text-md min-h-[4rem] border border-base-300">
                                        Then I took an arrow in the knee.
                                    </div>
                                </div>

                                <div className="flex flex-col gap-1 flex-1 relative min-h-[12rem]">
                                    <div className="flex justify-between items-end shrink-0">
                                        <label className="label pb-0"><span className="label-text font-bold">訳文 (Target)</span></label>
                                        <span className="text-xs text-info font-bold mb-1">AI提案 (Gemini 1.5 Pro)</span>
                                    </div>
                                    <textarea className="textarea textarea-bordered textarea-primary flex-1 text-md leading-relaxed p-3" defaultValue={`そして膝に矢を受けてしまってな。`}></textarea>
                                </div>

                                {/* アクション */}
                                <div className="flex justify-end gap-2 mt-4 pt-4 border-t border-base-300 shrink-0">
                                    <button className="btn btn-outline btn-error btn-sm mr-auto">翻訳を除外</button>
                                    <button className="btn btn-outline btn-primary btn-sm">再生成</button>
                                    <button className="btn btn-success btn-sm text-white" onClick={() => setActiveTab(4)}>確定して次へ (Enter)</button>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* 5. エクスポートパネル (Export) */}
                    <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${activeTab === 4 ? 'flex' : 'hidden'}`}>
                        <div className="alert alert-success shadow-sm text-white shrink-0">
                            <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <span>翻訳フェーズが完了しました。成果物をエクスポートできます。</span>
                        </div>
                        <div className="card bg-base-100 border border-base-200 flex-1 shadow-sm">
                            <div className="card-body items-center text-center justify-center gap-6">
                                <h2 className="card-title text-2xl font-bold">エクスポート設定</h2>

                                <div className="form-control w-full max-w-md">
                                    <label className="label">
                                        <span className="label-text font-bold">出力形式</span>
                                    </label>
                                    <select className="select select-bordered w-full">
                                        <option>xTranslator 用 XML (.xml) - 推奨</option>
                                        <option>テキスト一覧 (.csv)</option>
                                        <option>JSON形式 (.json)</option>
                                    </select>
                                </div>

                                <div className="form-control w-full max-w-md text-left bg-base-200 p-4 rounded-xl border border-base-300">
                                    <label className="label cursor-pointer justify-start gap-4">
                                        <input type="checkbox" defaultChecked className="checkbox checkbox-primary" />
                                        <span className="label-text">未翻訳の行は原文を保持する</span>
                                    </label>
                                    <label className="label cursor-pointer justify-start gap-4 mt-2">
                                        <input type="checkbox" className="checkbox checkbox-primary" />
                                        <span className="label-text">翻訳フラグを「Validated (緑)」に設定する</span>
                                    </label>
                                </div>

                                <div className="card-actions mt-6 w-full max-w-md flex flex-col gap-4">
                                    <button className="btn btn-primary btn-lg w-full">成果物をダウンロード</button>
                                    <button className="btn btn-outline btn-sm w-full">プロジェクトをアーカイブ</button>
                                </div>
                                <div className="text-sm text-gray-400 mt-4">
                                    最終更新: 2026/02/26 23:09
                                </div>
                            </div>
                        </div>
                    </div>

                </div>
            </div>
        </div>
    );
};

export default TranslationFlow;
