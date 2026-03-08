import { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useShallow } from 'zustand/react/shallow';
import { useTaskStore } from '../../../store/taskStore';
import type { FrontendTask } from '../../../types/task';

/**
 * ダッシュボード画面で必要なジョブ一覧 state と操作を返す。
 */
export function useDashboard() {
    const tasks = useTaskStore(useShallow((state) => Object.values(state.tasks)));
    const resumeTask = useTaskStore((state) => state.resumeTask);
    const cancelTask = useTaskStore((state) => state.cancelTask);
    const navigate = useNavigate();

    const sortedTasks = useMemo(
        () => [...tasks].sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()),
        [tasks],
    );

    const handleTaskClick = (task: FrontendTask) => {
        const navigationState = { state: { taskId: task.id, phase: task.phase } };

        switch (task.type) {
            case 'dictionary_build':
                navigate('/dictionary', navigationState);
                return;
            case 'persona_extraction':
                navigate('/master_persona', navigationState);
                return;
            case 'translation_project':
                navigate('/translation_flow', navigationState);
                return;
        }
    };

    const handleResumeClick = (task: FrontendTask) => {
        if (task.type === 'persona_extraction') {
            navigate('/master_persona', { state: { taskId: task.id, phase: task.phase, resumeFromDashboard: true } });
            return;
        }

        void resumeTask(task);
    };

    return {
        sortedTasks,
        cancelTask,
        handleTaskClick,
        handleResumeClick,
    };
}
