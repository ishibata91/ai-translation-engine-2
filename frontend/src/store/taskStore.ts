import { create } from 'zustand'
import { FrontendTask } from '../types/task'
import * as Events from '../wailsjs/wailsjs/runtime/runtime'
import { GetActiveTasks, ResumeTask, CancelTask } from '../wailsjs/wailsjs/go/task/Bridge'

interface TaskState {
    tasks: Record<string, FrontendTask>;
    isLoading: boolean;

    setTasks: (tasks: FrontendTask[]) => void;
    updateTask: (task: FrontendTask) => void;
    removeTask: (taskId: string) => void;

    fetchActiveTasks: () => Promise<void>;
    resumeTask: (taskId: string) => Promise<void>;
    cancelTask: (taskId: string) => Promise<void>;
}

export const useTaskStore = create<TaskState>((set, get) => ({
    tasks: {},
    isLoading: false,

    setTasks: (tasks) => {
        const taskMap: Record<string, FrontendTask> = {};
        tasks.forEach(t => {
            taskMap[t.id] = t;
        });
        set({ tasks: taskMap });
    },

    updateTask: (task) => {
        set((state) => ({
            tasks: {
                ...state.tasks,
                [task.id]: task
            }
        }));
    },

    removeTask: (taskId) => {
        set((state) => {
            const newTasks = { ...state.tasks };
            delete newTasks[taskId];
            return { tasks: newTasks };
        });
    },

    fetchActiveTasks: async () => {
        set({ isLoading: true });
        try {
            const tasks = (await GetActiveTasks() as any) as FrontendTask[];
            get().setTasks(tasks);
        } catch (error) {
            console.error('Failed to fetch active tasks:', error);
        } finally {
            set({ isLoading: false });
        }
    },

    resumeTask: async (taskId) => {
        try {
            await ResumeTask(taskId);
        } catch (error) {
            console.error('Failed to resume task:', error);
            throw error;
        }
    },

    cancelTask: async (taskId) => {
        try {
            await CancelTask(taskId);
        } catch (error) {
            console.error('Failed to cancel task:', error);
            throw error;
        }
    }
}));

// Initialize event listeners
export const initTaskListeners = () => {
    Events.EventsOn('task:updated', (task: FrontendTask) => {
        useTaskStore.getState().updateTask(task);
    });

    Events.EventsOn('task:phase_completed', (payload: { taskId: string, phase: string, summary: any }) => {
        console.log('Phase completed:', payload);
        // Additional handling can be added here (e.g., refetching data related to the task)
    });
};
