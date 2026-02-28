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

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>長文の書籍や連続する会話シーンの要約を生成し、翻訳時のコンテキストとして利用します。</span>
                <div className="flex-none">
                    <button className="btn btn-sm btn-primary">要約の生成開始</button>
                </div>
            </div>

            <div className="shrink-0">
                <ModelSettings title="要約生成モデル設定" />
            </div>

            <div className="flex-1 flex flex-col min-h-0">
                <DataTable
                    columns={SUMMARY_COLUMNS}
                    data={SUMMARY_DATA}
                    selectedRowId={selectedRowId}
                    onRowSelect={(_row, id) => setSelectedRowId(id)}
                />
            </div>

            <div className="flex justify-end gap-2 shrink-0">
                <button className="btn btn-primary" onClick={onNext}>要約を確定して次へ</button>
            </div>
        </div>
    );
};
