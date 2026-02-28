import React from 'react';

interface SummaryPanelProps {
    isActive: boolean;
    onNext: () => void;
}

export const SummaryPanel: React.FC<SummaryPanelProps> = ({ isActive, onNext }) => {
    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
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
                <button className="btn btn-primary" onClick={onNext}>要約を確定して次へ</button>
            </div>
        </div>
    );
};
