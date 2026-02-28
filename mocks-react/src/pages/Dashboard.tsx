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
                                <tr className="hover">
                                    <td>LotD.esp</td>
                                    <td><span className="badge badge-secondary badge-outline badge-sm">翻訳中</span></td>
                                    <td className="w-1/3">
                                        <div className="flex items-center gap-2">
                                            <progress className="progress progress-secondary flex-1" value="40" max="100"></progress>
                                            <span className="text-xs min-w-[32px]">40%</span>
                                        </div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs">詳細</button></td>
                                </tr>
                                <tr className="hover">
                                    <td>SkyUI.esp</td>
                                    <td><span className="badge badge-primary badge-outline badge-sm">解析中</span></td>
                                    <td className="w-1/3">
                                        <div className="flex items-center gap-2">
                                            <progress className="progress progress-primary flex-1" value="70" max="100"></progress>
                                            <span className="text-xs min-w-[32px]">70%</span>
                                        </div>
                                    </td>
                                    <td><button className="btn btn-ghost btn-xs">詳細</button></td>
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
