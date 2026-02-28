import React from 'react';
import ModelSettings from '../ModelSettings';

interface TranslationPanelProps {
    isActive: boolean;
    onNext: () => void;
}

export const TranslationPanel: React.FC<TranslationPanelProps> = ({ isActive, onNext }) => {
    return (
        <div className={`tab-content-panel flex-col gap-4 h-full min-h-0 overflow-hidden ${isActive ? 'flex' : 'hidden'}`}>
            <div className="shrink-0">
                <ModelSettings title="翻訳モデル設定" />
            </div>

            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden">
                {/* 左ペイン: 翻訳対象リスト */}
                <div className="w-1/2 flex flex-col border rounded-xl bg-base-100 overflow-hidden">
                    <div className="p-3 border-b flex flex-col gap-2 bg-base-200 shrink-0">
                        <div className="text-sm font-bold">レコード一覧 (全 1,200 件)</div>
                        <div className="flex gap-2">
                            <input type="text" placeholder="原文・訳文を検索..." className="input input-bordered w-full input-sm" />
                            <select className="select select-bordered select-sm">
                                <option>未翻訳 (800)</option>
                                <option>AI翻訳済み (150)</option>
                                <option>確認済み (15)</option>
                                <option>翻訳済み (235)</option>
                                <option>除外済み (0)</option>
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
                                    <td><span className="badge badge-warning badge-xs text-white">確認済み</span></td>
                                </tr>
                                <tr className="hover cursor-pointer">
                                    <td>0001A2B3</td>
                                    <td className="truncate max-w-[200px]">I used to be an adventurer like you.</td>
                                    <td><span className="badge badge-success badge-xs">AI翻訳済み</span></td>
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
                        <span className="font-bold">詳細</span>
                    </div>

                    <div className="p-4 flex flex-col gap-4 flex-1 overflow-y-auto">
                        <div className="flex justify-between items-center text-sm shrink-0">
                            <span className="badge badge-outline">Record: 0001A2B4</span>
                            <div className="flex items-center gap-2">
                                <span className="font-bold text-base-content/70">状態:</span>
                                <div className="join">
                                    <button className="join-item btn btn-xs">未</button>
                                    <button className="join-item btn btn-xs">AI</button>
                                    <button className="join-item btn btn-xs btn-warning text-white">確認済</button>
                                    <button className="join-item btn btn-xs">翻訳済</button>
                                    <button className="join-item btn btn-xs btn-error btn-outline">除外</button>
                                </div>
                            </div>
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
                            </div>
                            <textarea className="textarea textarea-bordered textarea-primary flex-1 text-md leading-relaxed p-3" defaultValue={`そして膝に矢を受けてしまってな。`}></textarea>
                        </div>

                        {/* アクション */}
                        <div className="flex justify-end gap-2 mt-4 pt-4 border-t border-base-300 shrink-0">
                            <div className="tooltip tooltip-top" data-tip="モデル設定を引き継いで翻訳 (バッチモードは無視)">
                                <button className="btn btn-secondary btn-sm text-white">単体翻訳リクエスト</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex justify-end gap-2 shrink-0">
                <button className="btn btn-primary" onClick={onNext}>確定して次へ</button>
            </div>
        </div>
    );
};
