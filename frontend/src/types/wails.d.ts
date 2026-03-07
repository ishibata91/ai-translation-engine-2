declare module '*wailsjs/runtime/runtime' {
  export function EventsOn(eventName: string, callback: (payload: any) => void): () => void;
  export function EventsOff(eventName: string): void;
}

declare module '*wailsjs/go/config/ConfigService' {
  export function ConfigGetAll(namespace: string): Promise<Record<string, string>>;
  export function ConfigSet(namespace: string, key: string, value: string): Promise<void>;
  export function UIStateGetJSON(namespace: string, key: string): Promise<string>;
  export function UIStateSetJSON(namespace: string, key: string, value: unknown): Promise<void>;
}

declare module '*wailsjs/go/modelcatalog/ModelCatalogService' {
  export function ListModels(input: {
    namespace: string;
    provider: string;
    endpoint: string;
    apiKey: string;
  }): Promise<Array<{ id: string; display_name?: string }>>;
}

declare module '*wailsjs/go/main/App' {
  export function SelectJSONFile(): Promise<string>;
  export function SelectFiles(): Promise<string[]>;
  export function DictGetSources(): Promise<any[]>;
  export function DictStartImport(filePath: string): Promise<number>;
  export function DictGetEntriesPaginated(sourceID: number, query: string, filters: Record<string, string>, page: number, pageSize: number): Promise<any>;
  export function DictSearchAllEntriesPaginated(query: string, filters: Record<string, string>, page: number, pageSize: number): Promise<any>;
  export function DictUpdateEntry(entry: any): Promise<void>;
  export function DictDeleteEntry(id: number): Promise<void>;
  export function DictDeleteSource(id: number): Promise<void>;
}

declare module '*wailsjs/go/task/Bridge' {
  export function StartMasterPersonTask(input: { source_json_path: string; overwrite_existing?: boolean }): Promise<string>;
  export function ResumeTask(taskID: string): Promise<void>;
  export function CancelTask(taskID: string): Promise<void>;
  export function GetTaskRequestState(taskID: string): Promise<{ total?: number; completed?: number; failed?: number; canceled?: number }>;
  export function GetAllTasks(): Promise<any[]>;
  export function GetActiveTasks(): Promise<any[]>;
}
