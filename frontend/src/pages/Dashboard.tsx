import React from 'react';

const Dashboard: React.FC = () => {
    return (
        <div className="flex flex-col w-full p-4 gap-4">
            {/* ヘッダー部分 */}
            <div className="navbar bg-base-100 rounded-box shadow-sm px-4">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">ダッシュボード (Dashboard)</span>
                </div>
            </div>

            {/* 進行中のジョブ */}
            <div className="card bg-base-100 shadow-sm border border-base-200 flex-1">
                <div className="card-body">
                    <h2 className="card-title text-base">進行中のジョブ</h2>
                    <div className="overflow-x-auto">
                        <table className="table w-full">
                            <thead>
                                <tr>
                                    <th>Mod名</th>
                                    <th>フェーズ</th>
                                    <th>進捗</th>
                                    <th>アクション</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td colSpan={4} className="text-center text-base-content/50 py-4">
                                        進行中のジョブはありません
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
