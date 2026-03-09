declare module '*wailsjs/runtime/runtime' {
  export function EventsOn<TPayload = unknown>(eventName: string, callback: (payload: TPayload) => void): () => void;
  export function EventsOff(eventName: string): void;
}

declare module '*wailsjs/go/controller/ConfigController' {
  export function ConfigGetAll(namespace: string): Promise<Record<string, string>>;
  export function ConfigSet(namespace: string, key: string, value: string): Promise<void>;
  export function UIStateGetJSON(namespace: string, key: string): Promise<string>;
  export function UIStateSetJSON(namespace: string, key: string, value: unknown): Promise<void>;
}

declare module '*wailsjs/go/controller/ModelCatalogController' {
  export function ListModels(input: {
    namespace: string;
    provider: string;
    endpoint: string;
    apiKey: string;
  }): Promise<Array<{ id: string; display_name?: string }>>;
}

declare module '*wailsjs/go/controller/FileDialogController' {
  export function SelectJSONFile(): Promise<string>;
  export function SelectFiles(): Promise<string[]>;
}

declare module '*wailsjs/go/controller/DictionaryController' {
  export function DictGetSources(): Promise<unknown[]>;
  export function DictStartImport(filePath: string): Promise<number>;
  export function DictGetEntriesPaginated(sourceID: number, query: string, filters: Record<string, string>, page: number, pageSize: number): Promise<unknown>;
  export function DictSearchAllEntriesPaginated(query: string, filters: Record<string, string>, page: number, pageSize: number): Promise<unknown>;
  export function DictUpdateEntry(entry: unknown): Promise<void>;
  export function DictDeleteEntry(id: number): Promise<void>;
  export function DictDeleteSource(id: number): Promise<void>;
}

declare module '*wailsjs/go/controller/PersonaController' {
  export function ListNPCs(): Promise<unknown[]>;
  export function ListDialoguesByPersonaID(personaID: number): Promise<unknown[]>;
}

declare module '*wailsjs/go/controller/TaskController' {
  export function ResumeTask(taskID: string): Promise<void>;
  export function CancelTask(taskID: string): Promise<void>;
  export function GetAllTasks(): Promise<unknown[]>;
  export function GetActiveTasks(): Promise<unknown[]>;
}

declare module '*wailsjs/go/controller/PersonaTaskController' {
  export function StartMasterPersonTask(input: { source_json_path: string; overwrite_existing?: boolean }): Promise<string>;
  export function ResumeTask(taskID: string): Promise<void>;
  export function CancelTask(taskID: string): Promise<void>;
  export function ResumeMasterPersonaTask(taskID: string): Promise<void>;
  export function GetTaskRequestState(taskID: string): Promise<{ total?: number; completed?: number; failed?: number; canceled?: number }>;
  export function GetTaskRequests(taskID: string): Promise<unknown[]>;
  export function GetAllTasks(): Promise<unknown[]>;
}
