import React from 'react';

interface ExportPanelProps {
    isActive: boolean;
}

export const ExportPanel: React.FC<ExportPanelProps> = ({ isActive }) => {
    return (
        <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${isActive ? 'flex' : 'hidden'}`}>
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
                            <option>xTranslator 用 XML (.xml)</option>
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
                    </div>
                    <div className="text-sm text-gray-400 mt-4">
                        最終更新: 2026/02/26 23:09
                    </div>
                </div>
            </div>
        </div>
    );
};
