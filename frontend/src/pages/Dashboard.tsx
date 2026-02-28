import React from 'react';
import { useTaskStore } from '../store/taskStore';
import { useNavigate } from 'react-router-dom';
import { FrontendTask } from '../types/task';

const Dashboard: React.FC = () => {
    const tasks = useTaskStore(state => Object.values(state.tasks));
    const resumeTask = useTaskStore(state => state.resumeTask);
    const cancelTask = useTaskStore(state => state.cancelTask);
    const navigate = useNavigate();

    const handleTaskClick = (task: FrontendTask) => {
        // Phase based routing
        const navigationState = { state: { taskId: task.id, phase: task.phase } };
        switch (task.type) {
            case 'dictionary_build':
                navigate('/dictionary', navigationState);
                break;
            case 'persona_extraction':
                navigate('/master_persona', navigationState);
                break;
            case 'translation_project':
                navigate('/translation_flow', navigationState);
                break;
        }
    };

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
                                    <th>ジョブ名</th>
                                    <th>タイプ</th>
                                    <th>フェーズ</th>
                                    <th>進捗</th>
                                    <th>ステータス</th>
                                    <th>アクション</th>
                                </tr>
                            </thead>
                            <tbody>
                                {tasks.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="text-center text-base-content/50 py-4">
                                            進行中のジョブはありません
                                        </td>
                                    </tr>
                                ) : (
                                    tasks.map(task => (
                                        <tr key={task.id} className="hover cursor-pointer" onClick={() => handleTaskClick(task)}>
                                            <td>{task.name}</td>
                                            <td>
                                                <div className="badge badge-ghost badge-sm">{task.type}</div>
                                            </td>
                                            <td>{task.phase}</td>
                                            <td>
                                                <div className="flex items-center gap-2">
                                                    <progress
                                                        className={`progress w-20 ${task.status === 'failed' ? 'progress-error' : 'progress-primary'}`}
                                                        value={task.progress}
                                                        max="100"
                                                    ></progress>
                                                    <span className="text-xs font-mono">{task.progress.toFixed(1)}%</span>
                                                </div>
                                            </td>
                                            <td>
                                                <div className={`badge badge-sm ${task.status === 'running' ? 'badge-primary' :
                                                    task.status === 'paused' ? 'badge-warning' :
                                                        task.status === 'failed' ? 'badge-error' : 'badge-ghost'
                                                    }`}>
                                                    {task.status}
                                                </div>
                                            </td>
                                            <td onClick={(e) => e.stopPropagation()}>
                                                <div className="flex gap-2">
                                                    {task.status === 'paused' && (
                                                        <button
                                                            className="btn btn-xs btn-success"
                                                            onClick={() => resumeTask(task.id)}
                                                        >再開</button>
                                                    )}
                                                    {task.status === 'running' && (
                                                        <button
                                                            className="btn btn-xs btn-outline btn-error"
                                                            onClick={() => cancelTask(task.id)}
                                                        >停止</button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
