import React, { useState } from 'react';
import GridEditor from '../GridEditor';
import type { GridColumnDef } from '../GridEditor';
import ModelSettings from '../ModelSettings';

interface TerminologyPanelProps {
    isActive: boolean;
    onNext: () => void;
}

interface TerminologyRow {
    type: string;
    source: string;
    target: string;
    status: string;
}

const mockColumns: GridColumnDef<TerminologyRow>[] = [
    { key: 'type', header: '種別', widthClass: 'w-1/6' },
    { key: 'source', header: '翻訳対象 (Source)', widthClass: 'w-2/6' },
    { key: 'target', header: '結果 (Result)', editable: true, widthClass: 'w-2/6' },
    { key: 'status', header: '状態 / ロジック', widthClass: 'w-1/6' },
];

const mockInitialData: TerminologyRow[] = [
    { type: 'LCTN:FULL', source: 'Whiterun', target: 'ホワイトラン', status: '辞書・完全一致' },
    { type: 'NPC_:FULL', source: 'Jon Battle-Born', target: 'ジョン・バトルボーン', status: '辞書・部分一致' },
    { type: 'WEAP:FULL', source: 'Skyforge Steel Sword', target: 'スカイフォージの鋼鉄の剣', status: 'LLM翻訳・完了' },
    { type: 'NPC_:SHRT', source: 'Ulfric', target: '翻訳中...', status: 'LLM取得中...' },
];

export const TerminologyPanel: React.FC<TerminologyPanelProps> = ({ isActive, onNext }) => {
    const [isTranslating, setIsTranslating] = useState(true);
    const [isGridDirty, setIsGridDirty] = useState(false);
    const [showConfirmModal, setShowConfirmModal] = useState(false);

    const handleNextClick = () => {
        if (isGridDirty) {
            setShowConfirmModal(true);
        } else {
            onNext();
        }
    };

    const handleConfirmProceed = () => {
        setShowConfirmModal(false);
        onNext();
    };

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>Mod内から抽出された固有名詞や特殊用語を翻訳し、Mod専用辞書を構築・保存しています。</span>
            </div>

            <div className="shrink-0">
                <ModelSettings title="用語翻訳モデル設定" />
            </div>



            {/* 翻訳結果リスト */}
            <div className="flex-1 relative border rounded-xl bg-base-100 overflow-hidden flex flex-col">
                <GridEditor
                    title="用語一覧"
                    columns={mockColumns}
                    initialData={mockInitialData}
                    onSave={() => { setIsGridDirty(false); }}
                    onDirtyChange={setIsGridDirty}
                />

                {/* 翻訳実行中のオーバーレイ (エディターの操作をブロック) */}
                {isTranslating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <span className="font-bold text-lg text-base-content/70">自動翻訳を実行中...</span>
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: TerminologyTranslation ({isTranslating ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isTranslating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={() => setIsTranslating(!isTranslating)}
                    >
                        {isTranslating ? '一時停止' : '再開'}
                    </button>
                    <button
                        className="btn btn-primary btn-sm"
                        onClick={handleNextClick}
                        disabled={isTranslating}
                    >
                        用語を確定して次へ
                    </button>
                </div>
            </div>

            {/* 未保存警告モーダル */}
            {showConfirmModal && (
                <div className="modal modal-open">
                    <div className="modal-box">
                        <h3 className="font-bold text-lg text-warning">未保存の変更があります</h3>
                        <p className="py-4">表内に保存されていない編集内容が存在します。保存せずに次のステップへ進みますか？</p>
                        <div className="modal-action">
                            <button className="btn" onClick={() => setShowConfirmModal(false)}>キャンセル</button>
                            <button className="btn btn-warning" onClick={handleConfirmProceed}>保存せずに進む</button>
                        </div>
                    </div>
                    <div className="modal-backdrop" onClick={() => setShowConfirmModal(false)}></div>
                </div>
            )}
        </div>
    );
};
