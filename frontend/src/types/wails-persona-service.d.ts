declare module '*wailsjs/go/controller/PersonaController' {
  export function ListNPCs(): Promise<unknown[]>;
  export function ListDialoguesByPersonaID(personaID: number): Promise<unknown[]>;
}
