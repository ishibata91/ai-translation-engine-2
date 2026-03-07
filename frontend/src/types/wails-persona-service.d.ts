declare module '*wailsjs/go/persona/Service' {
  export function ListNPCs(): Promise<unknown[]>;
  export function ListDialoguesByPersonaID(personaID: number): Promise<unknown[]>;
}
