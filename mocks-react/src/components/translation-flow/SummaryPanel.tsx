import React, { useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../ModelSettings';
import DataTable from '../DataTable';

interface SummaryPanelProps {
    isActive: boolean;
    onNext: () => void;
}

interface SummaryRow {
    type: string;
    typeColor: string;
    target: string;
    status: string;
    statusColor: string;
    content: React.ReactNode;
}

const SUMMARY_DATA: SummaryRow[] = [
    { type: 'Book', typeColor: 'badge-primary', target: 'The Lusty Argonian Maid, v1', status: '完了', statusColor: 'badge-success', content: 'アルゴニアンのメイドであるリフトスと、主人のクラシウス・キュリオの際どい会話劇。比喩表現が多用される。' },
    { type: 'Dialog', typeColor: 'badge-secondary', target: 'MQ101_UlfricExecution', status: '完了', statusColor: 'badge-success', content: 'ヘルゲンでの処刑シーン。帝国軍によるストームクローク兵の処刑と、突然のドラゴンの襲撃。緊迫した雰囲気。' },
    { type: 'Dialog', typeColor: 'badge-secondary', target: 'MQ102_RiverwoodArrive', status: '未生成', statusColor: 'badge-ghost', content: <span className="text-gray-400 italic">（未生成）</span> },
];

const SUMMARY_COLUMNS: ColumnDef<SummaryRow, any>[] = [
    {
        accessorKey: 'type',
        header: '種別',
        cell: (info) => <span className={`badge ${info.row.original.typeColor} badge-sm text-white`}>{info.getValue<string>()}</span>,
    },
    {
        accessorKey: 'target',
        header: '対象レコード/シーン',
    },
    {
        accessorKey: 'status',
        header: '状態',
        cell: (info) => <span className={`badge ${info.row.original.statusColor} badge-sm`}>{info.getValue<string>()}</span>,
    },
    {
        accessorKey: 'content',
        header: '要約内容',
        cell: (info) => <div className="text-sm">{info.getValue<React.ReactNode>()}</div>,
    }
];

export const SummaryPanel: React.FC<SummaryPanelProps> = ({ isActive, onNext }) => {
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [isGenerating, setIsGenerating] = useState<boolean>(true);

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>長文の書籍や連続する会話シーンの要約を生成し、翻訳時のコンテキストとして利用します。</span>
            </div>

            <div className="shrink-0">
                <ModelSettings title="要約生成モデル設定" />
            </div>

            <div className="flex-1 flex flex-col min-h-0 relative">
                <DataTable
                    columns={SUMMARY_COLUMNS}
                    data={SUMMARY_DATA}
                    selectedRowId={selectedRowId}
                    onRowSelect={(_row, id) => setSelectedRowId(id)}
                />

                {isGenerating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl border border-base-200">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <span className="font-bold text-lg text-base-content/70">要約を自動生成中...</span>
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: SummaryGeneration ({isGenerating ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isGenerating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={() => setIsGenerating(!isGenerating)}
                    >
                        {isGenerating ? '一時停止' : '再開'}
                    </button>
                    <button className="btn btn-primary btn-sm" onClick={onNext} disabled={isGenerating}>
                        要約を確定して次へ
                    </button>
                </div>
            </div>
        </div>
    );
};
