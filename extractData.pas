{ ========================================================================
   xEdit Data Extraction Script for translate_with_local_ai.py v2.0
   Purpose: Extract structured context from Skyrim ESM/ESP files
   Output: JSON file with dialogue groups, quests, items, NPCs, etc.
   ======================================================================== }

unit ExportData;

var
  // Output buffers
  dialogueList: TStringList;
  questList: TStringList;
  itemList: TStringList;
  magicList: TStringList;
  locationList: TStringList;
  cellList: TStringList;
  systemList: TStringList;
  messageList: TStringList;
  loadScreenList: TStringList;
  npcList: TStringList;
  processedNPCs: TStringList;

  
  // Dialogue Maps - Single Consolidated Map
  // Key = DialID. Object = TStringList.
  // Inner List Index 0: Header (DIAL JSON start)
  // Inner List Index 1..N: Responses (INFO JSON strings)
  dialDataMap: TStringList;
  
  // State
  targetFileName: string;
  
  // Ported from exportDialogue.pas
  lstRecursion: TList;
  InfoNPCID, InfoSPEAKER, InfoRACEID, InfoCONDITION: string;
  const fDebug = False;

// ===== UTILITY FUNCTIONS =====


function HexFormID(elem: IInterface): string;
begin
  if Assigned(elem) then
    Result := IntToHex(FixedFormID(elem), 8)
  else
    Result := '00000000';
end;



function GetElementValue(elem: IInterface; path: string): string;
begin
  Result := '';
  if not Assigned(elem) then Exit;
  Result := GetElementEditValues(elem, path);
end;

function GetElementValueByName(elem: IInterface; name: string): string;
var
  child: IInterface;
begin
  Result := '';
  child := ElementByName(elem, name);
  if Assigned(child) then
    Result := GetEditValue(child);
end;

function EscapeJSONString(const s: string): string;
var
  i: Integer;
  ch: Char;
  c: Integer;
  chStr: string;
begin
  Result := '';
  for i := 1 to Length(s) do begin
    chStr := Copy(s, i, 1);
    if Length(chStr) > 0 then
    begin
      ch := chStr[1];
      c := Ord(ch);
    end else c := 0;
    
    case c of
      34: Result := Result + '\"';  // "
      92: Result := Result + '\\';  // \
      47: Result := Result + '\/';  // /
      8:  Result := Result + '\b';
      9:  Result := Result + '\t';
      10: Result := Result + '\n';
      12: Result := Result + '\f';
      13: Result := Result + '\r';
    else
      if (c < 32) or (c > 127) then
        Result := Result + '\u' + IntToHex(c, 4)
      else
        Result := Result + ch;
    end;
  end;
end;

function JsonString(s: string): string;
begin
  Result := '"' + EscapeJSONString(s) + '"';
end;

function JsonField(key, value: string): string;
begin
  Result := '  "' + key + '": ' + value;
end;

function GetFileNameWithExtension(aFile: IwbFile): string;
var
  FileName: string;
  HasESL: boolean;
  HasESM: boolean;
  lower: string;
begin
  FileName := GetFileName(aFile);
  HasESL := GetIsESL(aFile);
  HasESM := GetIsESM(aFile);
  lower := LowerCase(FileName);

  if (Pos('.esl', lower) > 0) or (Pos('.esm', lower) > 0) or (Pos('.esp', lower) > 0) then
    Result := FileName
  else if HasESL then
    Result := FileName + '.esl'
  else if HasESM then
    Result := FileName + '.esm'
  else
    Result := FileName + '.esp';
end;

function GetMasterFileName(e: IInterface): string;
var
  recordFile: IwbFile;
begin
  if not Assigned(e) then
    Result := ''
  else
  begin
    recordFile := GetFile(e);
    if not Assigned(recordFile) then
      Result := ''
    else
      Result := GetFileNameWithExtension(recordFile);
  end;
end;

// ===== REORDERED NPC LOGIC =====
procedure ExtractNPC(npc: IInterface);
var
  npcID, npcName, race, voice, sexFlag, sex, class_, npcEntry, source: string;
begin
  npcID := HexFormID(npc);
  if processedNPCs.IndexOf(npcID) >= 0 then Exit;
  processedNPCs.Add(npcID);

  npcName := GetElementValue(npc, 'FULL');
  race := GetElementValue(npc, 'RNAM');
  voice := GetElementValue(npc, 'VTCK');
  class_ := GetElementValue(npc, 'CNAM');
  source := GetMasterFileName(npc);
  
  if (GetElementNativeValues(npc, 'ACBS\Flags') and 1) <> 0 then sex := 'Female' else sex := 'Male';
  
  if (npcName <> '') then
  begin
    npcEntry := '  "' + npcID + '": {' + #13#10 +
                '    "id": ' + JsonString(npcID) + ',' + #13#10 +
                '    "editor_id": ' + JsonString(GetElementEditValues(MasterOrSelf(npc), 'EDID')) + ',' + #13#10 +
                '    "type": "NPC_ FULL",' + #13#10 +
                '    "source": ' + JsonString(source) + ',' + #13#10 +
                '    "name": ' + JsonString(npcName) + ',' + #13#10 +
                '    "race": ' + JsonString(race) + ',' + #13#10 +
                '    "voice": ' + JsonString(voice) + ',' + #13#10 +
                '    "sex": ' + JsonString(sex) + ',' + #13#10 +
                '    "class": ' + JsonString(class_) + #13#10 +
                '  }';
    npcList.Add(npcEntry);
  end;
end;

procedure EnsureNPC(e: IInterface);
begin
  if not Assigned(e) then Exit;
  if Signature(e) = 'NPC_' then ExtractNPC(e);
end;

// ===== EXTRACTION LOGIC =====

function GetOrCreateDialList(dialID: string): TStringList;
var
  idx: integer;
  lst: TStringList;
begin
  idx := dialDataMap.IndexOf(dialID);
  if idx = -1 then
  begin
    lst := TStringList.Create;
    // Pre-fill header slot with empty to ensure index 0 is reserved
    lst.Add(''); 
    dialDataMap.AddObject(dialID, lst);
    Result := lst;
  end else
  begin
    if dialDataMap.Objects[idx] = nil then
    begin
         // Should not happen, but recover
         lst := TStringList.Create;
         lst.Add('');
         dialDataMap.Objects[idx] := lst;
         Result := lst;
    end else
    begin
         Result := TStringList(dialDataMap.Objects[idx]);
    end;
  end;
end;


procedure ExtractDialogue(dial: IInterface);
var
  dialID: string;
  playerText: string;
  questID: string;
  subtype: string;
  nam1: string;
  isServices: string;
  dialJSON: string;
  lst: TStringList;
begin
  dialID := HexFormID(dial);
  playerText := GetElementValue(dial, 'FULL');
  questID := GetElementValue(dial, 'QNAM');
  subtype := GetElementValue(dial, 'DATA\Subtype');
  nam1 := GetElementValue(dial, 'NAM1');
  
  if (subtype = 'Services') or (subtype = 'Training') or (subtype = 'Barter') or (subtype = 'Misc') then
     isServices := 'true'
  else
     isServices := 'false';
  
  dialJSON := '  {' + #13#10 +
              '  "id": ' + JsonString(dialID) + ',' + #13#10 +
              '  "editor_id": ' + JsonString(GetElementEditValues(MasterOrSelf(dial), 'EDID')) + ',' + #13#10 +
              '  "type": "DIAL FULL",' + #13#10 +
              '  "player_text": ' + JsonString(playerText) + ',' + #13#10 +
              '  "source": ' + JsonString(GetMasterFileName(dial)) + ',' + #13#10 +
              '  "quest_id": ' + JsonString(questID) + ',' + #13#10 +
              '  "is_services_branch": ' + isServices + ',' + #13#10 +
              '  "services_type": ' + JsonString(subtype) + ',';

  if nam1 <> '' then
     dialJSON := dialJSON + #13#10 + '  "nam1": ' + JsonString(nam1) + ',';
  
  lst := GetOrCreateDialList(dialID);
  lst[0] := dialJSON; // Set Header
end;

// ================= COPIED FROM exportDialogue.pas =================

// Forward declarations might be needed if not ordered correctly.
// GetRecordVoiceTypes calls GetRecordVoiceTypes2.
// GetAliasVoiceTypes calls GetRecordVoiceTypes.
// GetConditionsVoiceTypes calls GetRecordVoiceTypes & GetAliasVoiceTypes.
// InfoVoiceTypes calls GetRecordVoiceTypes & GetConditionsVoiceTypes & GetAliasVoiceTypes.
// Order: define GetRecordVoiceTypes2 first (simplest recursive), then others.

procedure GetRecordVoiceTypes2(e: IInterface; lstVoice: TStringList);
var
  sig: string;
  i: integer;
  ent, ents: IInterface;
begin
  if lstRecursion.IndexOf(FormID(e)) <> -1 then Exit
    else lstRecursion.Add(FormID(e));
  
  e := WinningOverride(e);
  // AddDebug('GetRecordVoiceType '+Name(e));

  sig := Signature(e);
  if (sig = 'REFR') or (sig = 'ACHR') then begin
    e := WinningOverride(BaseRecord(e));
    sig := Signature(e);
  end;
    
  if sig = 'VTYP' then begin
    lstVoice.AddObject(EditorID(e), e);
  end
  else if sig = 'NPC_' then begin
    if ElementExists(e, 'VTCK - Voice') then begin
      InfoNPCID := HexFormID(e);
      EnsureNPC(e);
      InfoRACEID := GetElementEditValues(e, 'RNAM - Race');
      ent := LinksTo(ElementByName(e, 'VTCK - Voice'));
      GetRecordVoiceTypes2(ent, lstVoice);
    end
    else if ElementExists(e, 'TPLT - Template') then begin
      ent := LinksTo(ElementByName(e, 'TPLT - Template'));
      GetRecordVoiceTypes2(ent, lstVoice);
    end;
  end
  else if sig = 'LVLN' then begin
    ents := ElementByName(e, 'Leveled List Entries');
    for i := 0 to Pred(ElementCount(ents)) do begin
      ent := ElementByIndex(ents, i);
      ent := LinksTo(ElementByPath(ent, 'LVLO\NPC'));
      GetRecordVoiceTypes2(ent, lstVoice);
    end;
  end
  else if sig = 'TACT' then begin
    ent := WinningOverride(LinksTo(ElementByName(e, 'VNAM - Voice Type')));
    if Signature(ent) = 'VTYP' then lstVoice.AddObject(EditorID(ent), ent);
  end
  else if sig = 'FLST' then begin
    ents := ElementByName(e, 'FormIDs');
    for i := 0 to Pred(ElementCount(ents)) do begin
      ent := LinksTo(ElementByIndex(ents, i));
      GetRecordVoiceTypes2(ent, lstVoice);
    end;
  end
  else if (sig = 'FACT') or (sig = 'CLAS') then begin
    for i := 0 to Pred(ReferencedByCount(e)) do begin
      ent := ReferencedByIndex(e, i);
      if Signature(ent) = 'NPC_' then
        GetRecordVoiceTypes2(ent, lstVoice);
    end;
  end;
end;

procedure GetRecordVoiceTypes(e: IInterface; lstVoice: TStringList);
begin
  lstRecursion.Clear;
  GetRecordVoiceTypes2(e, lstVoice);
end;



procedure GetAliasVoiceTypes(Quest: IInterface; aAlias: integer; lstVoice: TStringList);
var
  Aliases, Alias: IInterface;
  i, j: integer;
  lstLimit: TStringList;
begin
  Quest := WinningOverride(Quest);
  // AddDebug('GetAliasVoiceTypes ' + Name(Quest) + ' -- ' + IntToStr(aAlias));
  
  Aliases := ElementByName(Quest, 'Aliases');
  for i := 0 to Pred(ElementCount(Aliases)) do begin
    Alias := ElementByIndex(Aliases, i);
    if GetNativeValue(ElementByIndex(Alias, 0)) <> aAlias then
      Continue;

    if ElementExists(Alias, 'ALFR - Forced Reference') then
      GetRecordVoiceTypes(LinksTo(ElementByName(Alias, 'ALFR - Forced Reference')), lstVoice)
    else if ElementExists(Alias, 'ALUA - Unique Actor') then
      GetRecordVoiceTypes(LinksTo(ElementByName(Alias, 'ALUA - Unique Actor')), lstVoice)
    else if ElementExists(Alias, 'External Alias Reference') then
      GetAliasVoiceTypes(
        LinksTo(ElementByPath(Alias, 'External Alias Reference\ALEQ - Quest')),
        GetElementNativeValues(Alias, 'External Alias Reference\ALEA - Alias'),
        lstVoice
      )
    else if ElementExists(Alias, 'Conditions') then
      GetConditionsVoiceTypes(ElementByName(Alias, 'Conditions'), lstVoice);
    
    if GetElementNativeValues(Alias, 'VTCK - Voice Types') <> 0 then begin
      if lstVoice.Count <> 0 then begin
        lstLimit := TStringList.Create;
        GetRecordVoiceTypes(LinksTo(ElementByName(Alias, 'VTCK - Voice Types')), lstLimit);
        for j := Pred(lstVoice.Count) downto 0 do
          if lstLimit.IndexOf(lstVoice[j]) = -1 then
            lstVoice.Delete(j);
        lstLimit.Free;
      end
      else
        GetRecordVoiceTypes(LinksTo(ElementByName(Alias, 'VTCK - Voice Types')), lstVoice);
    end;
    Break;
  end;
end;

procedure GetConditionsVoiceTypes(Conditions: IInterface; lstVoice: TStringList);
var
  Condition, Elem, Quest: IInterface;
  ConditionFunction: string;
  lstVoiceCondition: TStringList;
  i, j, Alias: integer;
  bFactionCondition, bGetIsID: Boolean;
begin
  // AddDebug('GetConditionsVoiceTypes ' + FullPath(Conditions));
  lstVoiceCondition := TStringList.Create; 
  lstVoiceCondition.Duplicates := dupIgnore; 
  lstVoiceCondition.Sorted := True;

  for i := 0 to Pred(ElementCount(Conditions)) do
    if GetElementEditValues(ElementByIndex(Conditions, i), 'CTDA\Function') = 'GetIsID' then begin
      bGetIsID := True;
      Break;
    end;

  for i := 0 to Pred(ElementCount(Conditions)) do begin
    Condition := ElementByIndex(Conditions, i);
    ConditionFunction := GetElementEditValues(Condition, 'CTDA\Function');
    
    if ConditionFunction = 'GetIsID' then begin
      InfoCONDITION := ConditionFunction;
      Elem := LinksTo(ElementByPath(Condition, 'CTDA\Base Object'));
      GetRecordVoiceTypes(Elem, lstVoiceCondition);
    end else
    if not bGetIsID then 
    if ConditionFunction = 'GetIsVoiceType' then begin
      InfoCONDITION := ConditionFunction;
      Elem := LinksTo(ElementByPath(Condition, 'CTDA\Voice Type'));
      GetRecordVoiceTypes(Elem, lstVoiceCondition);
    end
    else if ConditionFunction = 'GetIsAliasRef' then begin
      Alias := GetElementNativeValues(Condition, 'CTDA\Alias');
      Elem := ContainingMainRecord(Conditions);
      if Signature(Elem) = 'INFO' then begin
        Elem := LinksTo(ElementByName(Elem, 'Topic'));
        Quest := LinksTo(ElementByName(Elem, 'QNAM - Quest'));
      end
      else if Signature(Elem) = 'QUST' then
        Quest := Elem;
      GetAliasVoiceTypes(Quest, Alias, lstVoiceCondition);
    end else
    if ConditionFunction = 'GetInFaction' then begin
      if not bFactionCondition then begin
        bFactionCondition := True;
        Elem := LinksTo(ElementByPath(Condition, 'CTDA\Faction'));
        GetRecordVoiceTypes(Elem, lstVoiceCondition);
      end;
    end
    else if ConditionFunction = 'GetIsClass' then begin
      Elem := LinksTo(ElementByPath(Condition, 'CTDA\Class'));
      GetRecordVoiceTypes(Elem, lstVoiceCondition);
    end;
    
    if lstVoiceCondition.Count = 0 then
      Continue;
    
    lstVoice.AddStrings(lstVoiceCondition);
    
    if GetElementNativeValues(Condition, 'CTDA\Type') and 1 > 0 then
      lstVoice.AddStrings(lstVoiceCondition)
    else begin
      for j := Pred(lstVoice.Count) downto 0 do
        if lstVoiceCondition.IndexOf(lstVoice[j]) = -1 then
          lstVoice.Delete(j);
    end;
    lstVoiceCondition.Clear;
  end;
  lstVoiceCondition.Free;
end;

procedure InfoVoiceTypes(Info: IInterface; lstVoice: TStringList; QuestConditions: IInterface);
var
  Elem, Dialogue: IInterface;
  Conditions: IInterface;
  Scene, Actions, Action: IInterface;
  i, j, Alias: integer;
  bAliasFound: Boolean;
  lstVoiceQuestConditions: TStringList;
begin
  if not Assigned(lstVoice) then Exit; 
    
  lstVoiceQuestConditions := TStringList.Create; 
  lstVoiceQuestConditions.Duplicates := dupIgnore; 
  lstVoiceQuestConditions.Sorted := True;
  lstVoice.Clear;
  
  Elem := ElementByName(Info, 'ANAM - Speaker');
  if Assigned(Elem) then begin
    Elem := LinksTo(Elem);
    GetRecordVoiceTypes(Elem, lstVoice);
    InfoSPEAKER := HexFormID(Elem);
    EnsureNPC(Elem);
    lstVoiceQuestConditions.Free;
    Exit;
  end;
  
  if QuestConditions <> nil then
    GetConditionsVoiceTypes(QuestConditions, lstVoiceQuestConditions);
  if ElementExists(Info, 'Conditions') then
    GetConditionsVoiceTypes(ElementByName(Info, 'Conditions'), lstVoice);
  if (lstVoice.Count = 0) and (lstVoiceQuestConditions.Count <> 0) then
    lstVoice.AddStrings(lstVoiceQuestConditions);
  if (lstVoice.Count <> 0) and (lstVoiceQuestConditions.Count <> 0) then
  begin
    for i := Pred(lstVoice.Count) downto 0 do
      if lstVoiceQuestConditions.IndexOf(lstVoice[i]) = -1 then
        lstVoice.Delete(i);
    // if lstVoice.Count = 0 then AddMessage('Warning: Info conditions conflict with quest conditions: ' + Name(Info));   
  end;  
  lstVoiceQuestConditions.Free;
  
  if lstVoice.Count <> 0 then Exit;

  bAliasFound := False;
  Dialogue := LinksTo(ElementByName(Info, 'Topic'));
  if GetElementEditValues(Dialogue, 'DATA\Category') = 'Scene' then begin
    for i := Pred(ReferencedByCount(Dialogue)) downto 0 do begin
      Scene := ReferencedByIndex(Dialogue, i);
      if Signature(Scene) <> 'SCEN' then Continue;
      
      Actions := ElementByName(Scene, 'Actions');
      for j := 0 to Pred(ElementCount(Actions)) do begin
        Action := ElementByIndex(Actions, j);
        if (GetElementEditValues(Action, 'ANAM - Type') = 'Dialogue') and
           (GetElementNativeValues(Action, 'DATA - Topic') = GetLoadOrderFormID(Dialogue))
        then begin
          Alias := GetElementNativeValues(Action, 'ALID - Actor ID');
          GetAliasVoiceTypes(LinksTo(ElementByName(Scene, 'PNAM - Quest')), Alias, lstVoice);
          bAliasFound := True;
          Break;
        end;
      end;
      if bAliasFound then Break;
    end;
  end;
end;



procedure ExtractInfo(info: IInterface);
var
  infoID, infoText, speakerID, voiceType, previousID, infoItem: string;
  dialID: string;
  parentRec, questRec, questConditions, responses: IInterface;
  lst, lstVoice: TStringList;
  i: integer;
  // New variables
  infoPrompt, topicText, menuDisplayText: string;
  responseIndex, responseNum: string;
  firstResponse: IInterface;
  responseCount: integer;
begin
  // Reset Globals
  InfoNPCID := '';
  InfoSPEAKER := '';
  InfoRACEID := '';
  InfoCONDITION := '';

  // Resolve Parent DIAL
  parentRec := GetContainer(info);
  while Assigned(parentRec) and (Signature(parentRec) <> 'DIAL') and (ElementType(parentRec) <> etFile) do
  begin
    parentRec := GetContainer(parentRec);
  end;
     
  if not Assigned(parentRec) or (Signature(parentRec) <> 'DIAL') then 
  begin
       parentRec := LinksTo(ElementByName(info, 'Topic'));
       if not Assigned(parentRec) or (Signature(parentRec) <> 'DIAL') then
       begin
           AddMessage('Debug: Info ' + HexFormID(info) + ' could not resolve DIAL parent via traversal or Topic link.');
           Exit;
       end;
  end;
  
  dialID := HexFormID(parentRec);
  infoID := HexFormID(info);
  
  // --- Common Field Extraction (shared across all Responses) ---
  infoPrompt := GetElementValue(info, 'RNAM');
  topicText := GetElementValue(parentRec, 'FULL');
  menuDisplayText := infoPrompt;

  // --- Speaker & VoiceType Resolution Strategy: Ported from exportDialogue.pas ---
  lstVoice := TStringList.Create;
  lstVoice.Duplicates := dupIgnore;
  lstVoice.Sorted := True;
  
  // Try to get Quest Conditions if possible
  // DIAL -> QNAM (Quest) -> Quest Dialogue Conditions
  questConditions := nil;
  if ElementExists(parentRec, 'QNAM') then
  begin
      questRec := LinksTo(ElementByName(parentRec, 'QNAM'));
      if Assigned(questRec) and ElementExists(questRec, 'Quest Dialogue Conditions') then
         questConditions := ElementByPath(questRec, 'Quest Dialogue Conditions\Conditions');
  end;
  
  // Call Ported Function
  InfoVoiceTypes(info, lstVoice, questConditions);
  
  // Determine final SpeakerID and VoiceType
  // exportDialogue logic: 
  // Speaker = InfoNPCID (Derived) or InfoSPEAKER (Explicit)
  // VoiceType = lstVoice[v]
  
  // Priority: 1. InfoSPEAKER (Explicit ANAM), 2. InfoNPCID (Derived from Conditions/Voice)
  speakerID := InfoSPEAKER;
  if speakerID = '' then speakerID := InfoNPCID;
  
  // Join Voice Types
  voiceType := lstVoice.CommaText;
  
  lstVoice.Free;
  
  previousID := GetElementValue(info, 'PNAM');
  
  // Get the DialDataMap list once (shared across all Responses)
  lst := GetOrCreateDialList(dialID);
  
  // --- Process each Response individually ---
  if ElementExists(info, 'Responses') then
  begin
       responses := ElementByName(info, 'Responses');
       responseCount := ElementCount(responses);
       
       // Loop through each Response and create individual JSON objects
       for i := 0 to responseCount - 1 do
       begin
            // Reset per-Response variables
            infoText := '';
            responseIndex := '';
            
            // Get current Response
            firstResponse := ElementByIndex(responses, i);
            
            // Extract Response text (NAM1)
            infoText := GetElementEditValues(firstResponse, 'NAM1');
            
            // Extract Response Number (TRDT\Response number) if it exists
            if ElementExists(firstResponse, 'TRDT') then
            begin
                responseNum := GetElementEditValues(firstResponse, 'TRDT\Response number');
                if responseNum <> '' then
                    responseIndex := responseNum;
            end;
            
            // Generate JSON object for this Response (if text or speaker exists)
            if (infoText <> '') or (speakerID <> '') then
            begin
              infoItem := '    {' + #13#10 +
                          '      "id": ' + JsonString(infoID) + ',' + #13#10 +
                          '      "editor_id": ' + JsonString(GetElementEditValues(MasterOrSelf(info), 'EDID')) + ',' + #13#10 +
                          '      "type": "INFO NAM1",' + #13#10 +
                          '      "source": ' + JsonString(GetMasterFileName(info)) + ',' + #13#10 +
                          '      "text": ' + JsonString(infoText) + ',' + #13#10 +
                          '      "prompt": ' + JsonString(infoPrompt) + ',' + #13#10 +
                          '      "topic_text": ' + JsonString(topicText) + ',' + #13#10 +
                          '      "menu_display_text": ' + JsonString(menuDisplayText) + ',' + #13#10 +
                          '      "speaker_id": ' + JsonString(speakerID) + ',' + #13#10 +
                          '      "voicetype": ' + JsonString(voiceType) + ',' + #13#10;
              
              // Add index field conditionally (only if Response Number exists)
              if responseIndex <> '' then
                  infoItem := infoItem + '      "index": ' + responseIndex + ',' + #13#10;
              
              infoItem := infoItem + '      "previous_id": ' + JsonString(previousID) + #13#10 +
                          '    }';
                          
              lst.Add(infoItem);
            end;
       end;
  end
  else
  begin
       // Fallback: No Responses element, use INFO direct NAM1
       infoText := GetElementValue(info, 'NAM1');
       responseIndex := '';
       
       if (infoText <> '') or (speakerID <> '') then
       begin
         infoItem := '    {' + #13#10 +
                     '      "id": ' + JsonString(infoID) + ',' + #13#10 +
                     '      "editor_id": ' + JsonString(GetElementEditValues(MasterOrSelf(info), 'EDID')) + ',' + #13#10 +
                     '      "type": "INFO NAM1",' + #13#10 +
                     '      "source": ' + JsonString(GetMasterFileName(info)) + ',' + #13#10 +
                     '      "text": ' + JsonString(infoText) + ',' + #13#10 +
                     '      "prompt": ' + JsonString(infoPrompt) + ',' + #13#10 +
                     '      "topic_text": ' + JsonString(topicText) + ',' + #13#10 +
                     '      "menu_display_text": ' + JsonString(menuDisplayText) + ',' + #13#10 +
                     '      "speaker_id": ' + JsonString(speakerID) + ',' + #13#10 +
                     '      "voicetype": ' + JsonString(voiceType) + ',' + #13#10 +
                     '      "previous_id": ' + JsonString(previousID) + #13#10 +
                     '    }';
                     
         lst.Add(infoItem);
       end;
  end;
end;



procedure ExtractQuest(quest: IInterface);
var
  questID, questName, questType, questEntry: string;

  stages, objectives, stageItem, objItem: IInterface;
  stageIdx, stageLog, objIdx, objText: string;
  stagesJson, objectivesJson, sEntry, oEntry: string;
  i, j: integer;
  subRecord1, subRecord2, subRecord3, subRecord4: IInterface;
begin
  questID := HexFormID(quest);
  questName := GetElementValue(quest, 'FULL');
  questType := GetElementValue(quest, 'DNAM\Quest Type'); 
  
  stagesJson := '';
  if ElementExists(quest, 'Stages') then
  begin
    subRecord1 := ElementByName(quest, 'Stages');
    for i := 0 to ElementCount(subRecord1) - 1 do
    begin
      subRecord2 := ElementByIndex(subRecord1, i);
      subRecord3 := ElementByName(subRecord2, 'INDX - Stage Index');
      if ElementCount(subRecord3) > 0 then
          stageIdx := IntToStr(Integer(GetNativeValue(ElementByIndex(subRecord3, 0))))
      else
          stageIdx := IntToStr(Integer(GetNativeValue(subRecord3)));
      
      // AddMessage('Debug: Quest Stage ' + IntToStr(i) + ' INDX: "' + stageIdx + '"');
      
      if ElementExists(subRecord2, 'Log Entries') then
      begin
        subRecord3 := ElementByName(subRecord2, 'Log Entries');
        for j := 0 to ElementCount(subRecord3) - 1 do
        begin
             subRecord4 := ElementByIndex(subRecord3, j);
             stageLog := GetElementValue(subRecord4, 'CNAM');
             
             if stageLog <> '' then
             begin
                  sEntry := '    {' + #13#10 +
                            '      "index": ' + IntToStr(StrToIntDef(stageIdx, 0) + j) + ',' + #13#10 +
                            '      "type": "QUST CNAM",' + #13#10 +
                            '      "text": ' + JsonString(stageLog) + #13#10 +
                            '    }';
                  if stagesJson <> '' then stagesJson := stagesJson + ',' + #13#10;
                  stagesJson := stagesJson + sEntry;
             end;
        end;
      end;
    end;
  end;

  objectivesJson := '';
  if ElementExists(quest, 'Objectives') then
  begin
    subRecord1 := ElementByName(quest, 'Objectives');
    for i := 0 to ElementCount(subRecord1) - 1 do
    begin
         subRecord2 := ElementByIndex(subRecord1, i);
         objIdx := GetElementValue(subRecord2, 'QOBJ'); 
         objText := GetElementValue(subRecord2, 'NNAM'); 
         
         if objText <> '' then
         begin
              oEntry := '    {' + #13#10 +
                        '      "index": ' + JsonString(objIdx) + ',' + #13#10 +
                        '      "type": "QUST NNAM",' + #13#10 +
                        '      "text": ' + JsonString(objText) + #13#10 +
                        '    }';
              if objectivesJson <> '' then objectivesJson := objectivesJson + ',' + #13#10;
              objectivesJson := objectivesJson + oEntry;
         end;
    end;
  end;

  if questName <> '' then
  begin
    questEntry := '  {' + #13#10 +
                  JsonField('id', JsonString(questID)) + ',' + #13#10 +
                  JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(quest), 'EDID'))) + ',' + #13#10 +
                  JsonField('source', JsonString(GetMasterFileName(quest))) + ',' + #13#10 +
                  JsonField('name', JsonString(questName)) + ',' + #13#10 +
                  JsonField('type', JsonString('QUST'));

    if stagesJson <> '' then
      questEntry := questEntry + ',' + #13#10 +
                    '  "stages": [' + #13#10 + stagesJson + #13#10 + '  ]';
                    
    if objectivesJson <> '' then
      questEntry := questEntry + ',' + #13#10 +
                    '  "objectives": [' + #13#10 + objectivesJson + #13#10 + '  ]';
                    
    questEntry := questEntry + #13#10 + '  }';
    questList.Add(questEntry);
  end;
end;

procedure ExtractItem(item: IInterface);
var
  itemID, itemName, itemDesc, itemText, typeHint, sig, itemEntry: string;
begin
  sig := Signature(item);
  itemID := HexFormID(item);
  itemName := GetElementValue(item, 'FULL');
  itemDesc := GetElementValue(item, 'DESC');
  typeHint := '';
  itemText := '';
  
  if sig = 'WEAP' then typeHint := GetElementValue(item, 'DNAM\Animation Type')
  else if sig = 'ARMO' then typeHint := GetElementValue(item, 'BODT\Armor Type')
  else if sig = 'BOOK' then itemText := GetElementValue(item, 'DESC'); 
  
  if (itemName <> '') then
  begin
    itemEntry := '  {' + #13#10 +
                 JsonField('id', JsonString(itemID)) + ',' + #13#10 +
                 JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(item), 'EDID'))) + ',' + #13#10 +
                 JsonField('type', JsonString(sig + ' FULL')) + ',' + #13#10 +
                 JsonField('source', JsonString(GetMasterFileName(item))) + ',' + #13#10 +
                 JsonField('name', JsonString(itemName)) + ',' + #13#10 +
                 JsonField('description', JsonString(itemDesc));
    
    if typeHint <> '' then
       itemEntry := itemEntry + ',' + #13#10 + JsonField('type_hint', JsonString(typeHint));
       
    if itemText <> '' then
       itemEntry := itemEntry + ',' + #13#10 + JsonField('text', JsonString(itemText));
      
    itemEntry := itemEntry + #13#10 + '  }';
    itemList.Add(itemEntry);
  end;
end;

procedure ExtractMagic(magic: IInterface);
var
  magicID, magicName, magicDesc, sig, magicEntry: string;
begin
  sig := Signature(magic);
  magicID := HexFormID(magic);
  magicName := GetElementValue(magic, 'FULL');
  magicDesc := GetElementValue(magic, 'DESC');
  if magicDesc = '' then magicDesc := GetElementValue(magic, 'DNAM');
  
  if magicName <> '' then
  begin
    magicEntry := '  {' + #13#10 +
                  JsonField('id', JsonString(magicID)) + ',' + #13#10 +
                  JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(magic), 'EDID'))) + ',' + #13#10 +
                  JsonField('type', JsonString(sig + ' FULL')) + ',' + #13#10 +
                  JsonField('source', JsonString(GetMasterFileName(magic))) + ',' + #13#10 +
                  JsonField('name', JsonString(magicName)) + ',' + #13#10 +
                  JsonField('description', JsonString(magicDesc)) + #13#10 +
                  '  }';
    magicList.Add(magicEntry);
  end;
end;

procedure ExtractLocation(location: IInterface);
var
  locID, locName, parentID, locEntry, sig: string;
begin
  sig := Signature(location);
  locID := HexFormID(location);
  locName := GetElementValue(location, 'FULL');
  
  if sig = 'LCTN' then parentID := GetElementValue(location, 'PNAM')
  else if sig = 'WRLD' then parentID := GetElementValue(location, 'WNAM')
  else if sig = 'CELL' then parentID := GetElementValue(location, 'XLCN');
  
  if locName <> '' then
  begin
    locEntry := '  {' + #13#10 +
                JsonField('id', JsonString(locID)) + ',' + #13#10 +
                JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(location), 'EDID'))) + ',' + #13#10 +
                JsonField('type', JsonString(sig + ' FULL')) + ',' + #13#10 +
                JsonField('source', JsonString(GetMasterFileName(location))) + ',' + #13#10 +
                JsonField('name', JsonString(locName)) + ',' + #13#10 +
                JsonField('parent_id', JsonString(parentID)) + #13#10 +
                '  }';
    locationList.Add(locEntry);
  end;
end;

procedure ExtractMessage(message: IInterface);
var
  msgID, msgText, msgTitle, questID, msgEntry: string;
begin
  msgID := HexFormID(message);
  msgText := GetElementValue(message, 'DESC');
  msgTitle := GetElementValue(message, 'FULL');
  questID := GetElementValue(message, 'QNAM');
  
  if msgText <> '' then
  begin
    msgEntry := '  {' + #13#10 +
                JsonField('id', JsonString(msgID)) + ',' + #13#10 +
                JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(message), 'EDID'))) + ',' + #13#10 +
                JsonField('type', JsonString('MESG DESC')) + ',' + #13#10 +
                JsonField('source', JsonString(GetMasterFileName(message))) + ',' + #13#10 +
                JsonField('text', JsonString(msgText)) + ',' + #13#10 +
                JsonField('title', JsonString(msgTitle)) + ',' + #13#10 +
                JsonField('quest_id', JsonString(questID)) + #13#10 +
                '  }';
    messageList.Add(msgEntry);
  end;
end;

procedure ExtractLoadScreen(lscr: IInterface);
var
  lscrID, lscrText, lscrEntry: string;
begin
  lscrID := HexFormID(lscr);
  lscrText := GetElementValue(lscr, 'DESC');
  
  if lscrText <> '' then
  begin
    lscrEntry := '  {' + #13#10 +
                 JsonField('id', JsonString(lscrID)) + ',' + #13#10 +
                 JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(lscr), 'EDID'))) + ',' + #13#10 +
                 JsonField('type', JsonString('LSCR DESC')) + ',' + #13#10 +
                 JsonField('source', JsonString(GetMasterFileName(lscr))) + ',' + #13#10 +
                 JsonField('text', JsonString(lscrText)) + #13#10 +
                 '  }';
    loadScreenList.Add(lscrEntry);
  end;
end;

procedure ExtractPerk(perk: IInterface);
var
  perkID, perkName, perkDesc, perkEntry: string;
begin
  perkID := HexFormID(perk);
  perkName := GetElementValue(perk, 'FULL');
  perkDesc := GetElementValue(perk, 'DESC');
  
  if perkName <> '' then
  begin
    perkEntry := '  {' + #13#10 +
                 JsonField('id', JsonString(perkID)) + ',' + #13#10 +
                 JsonField('editor_id', JsonString(GetElementEditValues(MasterOrSelf(perk), 'EDID'))) + ',' + #13#10 +
                 JsonField('type', JsonString('PERK FULL')) + ',' + #13#10 +
                 JsonField('source', JsonString(GetMasterFileName(perk))) + ',' + #13#10 +
                 JsonField('name', JsonString(perkName)) + ',' + #13#10 +
                 JsonField('description', JsonString(perkDesc)) + #13#10 +
                 '  }';
    systemList.Add(perkEntry);
  end;
end;

// ===== MAIN PROCESSING =====


procedure AddDebug(aMsg: string);
begin
  if fDebug then
    AddMessage('DEBUG: ' + aMsg);
end;

function Initialize: integer;
begin
  dialogueList := TStringList.Create;
  dialogueList.Clear;
  questList := TStringList.Create;
  questList.Clear;
  itemList := TStringList.Create;
  itemList.Clear;
  magicList := TStringList.Create;
  magicList.Clear;
  locationList := TStringList.Create;
  locationList.Clear;
  cellList := TStringList.Create;
  cellList.Clear;
  systemList := TStringList.Create;
  systemList.Clear;
  messageList := TStringList.Create;
  messageList.Clear;
  loadScreenList := TStringList.Create;
  loadScreenList.Clear;
  
  npcList := TStringList.Create;
  npcList.Clear;
  npcList.Duplicates := dupIgnore; 
  npcList.Sorted := True; 
  
  // Initialize Single Map (Unsorted to safely use AddObject)
  dialDataMap := TStringList.Create;
  dialDataMap.Sorted := False; 
  
  // Ported
  lstRecursion := TList.Create;
  
  processedNPCs := TStringList.Create;
  processedNPCs.Sorted := True;
  processedNPCs.Duplicates := dupIgnore;
  
  targetFileName := '';
  AddMessage('[ExportData] Initialized.');
  Result := 0;
end;

function Process(e: IInterface): integer;
var
  sig: string;
begin
  if targetFileName = '' then
    targetFileName := GetFileNameWithExtension(GetFile(e));
  
  sig := Signature(e);
  // Verbose log for every record processed
  // AddMessage('Process: ' + sig + ' ' + HexFormID(e)); 
  
  if sig = 'DIAL' then ExtractDialogue(e)
  else if sig = 'INFO' then ExtractInfo(e)
  else if sig = 'NPC_' then ExtractNPC(e)
  else if sig = 'QUST' then ExtractQuest(e)
  else if sig = 'MESG' then ExtractMessage(e)
  else if sig = 'LSCR' then ExtractLoadScreen(e)
  else if sig = 'PERK' then ExtractPerk(e)
  else if (sig = 'WEAP') or (sig = 'ARMO') or (sig = 'AMMO') or
          (sig = 'ALCH') or (sig = 'INGR') or (sig = 'KEYM') or
          (sig = 'MISC') or (sig = 'LIGH') or (sig = 'CONT') or
          (sig = 'SLGM') or (sig = 'BOOK') then ExtractItem(e)
  else if (sig = 'SPEL') or (sig = 'MGEF') or (sig = 'ENCH') or
          (sig = 'SCRL') or (sig = 'SHOU') then ExtractMagic(e)
  else if (sig = 'LCTN') or (sig = 'WRLD') or (sig = 'CELL') then ExtractLocation(e);
  
  Result := 0;
end;


function ExtractJsonValue(const json, key: string): string;
var
  p, q: Integer;
  searchKey: string;
begin
  Result := '';
  searchKey := '"' + key + '": "';
  p := Pos(searchKey, json);
  if p > 0 then
  begin
    p := p + Length(searchKey);
    q := Pos('"', Copy(json, p, Length(json)));
    if q > 0 then
      Result := Copy(json, p, q - 1);
  end;
end;

procedure ApplyInfoSorting(var respList: TStringList);
var
  i, j, k: Integer;
  pnamMap: TStringList;
  sortedList: TStringList;
  json, id, pid: string;
  siblings: TStringList;
  queue: TStringList; // Using StringList as Queue
  currentID: string;
  processedCount: Integer;
begin
  if respList.Count <= 1 then Exit; // 0 or 1 item, no sort needed

  pnamMap := TStringList.Create;
  pnamMap.Sorted := True;
  pnamMap.Duplicates := dupIgnore;

  // Build Map: PrevID -> List of JSONs
  for i := 1 to respList.Count - 1 do // Index 0 is Header, skip
  begin
    json := respList[i];
    id := ExtractJsonValue(json, 'id');
    pid := ExtractJsonValue(json, 'previous_id');
    
    // Default empty pid to "00000000" if standard null handling needed, 
    // but here empty string is fine. ExtractInfo gives empty string if missing.
    
    j := pnamMap.IndexOf(pid);
    if j < 0 then
    begin
      siblings := TStringList.Create;
      pnamMap.AddObject(pid, siblings);
      siblings.Add(json);
    end else
    begin
      siblings := TStringList(pnamMap.Objects[j]);
      siblings.Add(json);
    end;
  end;

  sortedList := TStringList.Create;
  sortedList.Add(respList[0]); // Keep Header

  queue := TStringList.Create;
  
  // Find roots (pid = empty or 00000000 or not found in IDs?)
  // Simplified: Start with pid="" 
  // If no root found via "", scan for orphans? 
  // Assuming standard valid chains start with empty PNAM.
  
  // Note: Some INFOs might have '00000000'. ExtractJsonValue gives string content.
  // ExtractData.pas uses `GetElementValue` which returns '00000000' for null formlinks usually?
  // Let's check `HexFormID`. '00000000' is returned if nil.
  
  if pnamMap.IndexOf('') >= 0 then
     queue.AddStrings(TStringList(pnamMap.Objects[pnamMap.IndexOf('')]));
  
  if pnamMap.IndexOf('00000000') >= 0 then
     queue.AddStrings(TStringList(pnamMap.Objects[pnamMap.IndexOf('00000000')]));

  // Process chain
  // Queue logic: Depth First? 
  // Standard chain: A -> B -> C.
  // Queue: [A]. Pop A. Result [A]. Children of A -> [B]. Push B. Queue [B].
  // BFS or DFS? "Linked List" implies linear. 
  // If branching: A -> B, A -> C.
  // We process A. Then B and C.
  // Order of B vs C matching game? Game uses visual order in CK?
  // We just use file order (insertion order in siblings list).
  
  // Implemented as FIFO Queue for Breadth traversal of forks, 
  // but linear chain is depth-like.
  // Actually, for a pure chain, Queue behavior doesn't matter.
  // For forks, we just process them.
  
  // Wait, "Queue" usually implies BFS. 
  // Let's use a "Work List".
  
  i := 0;
  while i < queue.Count do
  begin
     json := queue[i];
     sortedList.Add(json);
     
     // Find children
     id := ExtractJsonValue(json, 'id');
     j := pnamMap.IndexOf(id);
     if j >= 0 then
     begin
         siblings := TStringList(pnamMap.Objects[j]);
         // Add children to the end of queue
         for k := 0 to siblings.Count - 1 do
             queue.Add(siblings[k]);
     end;
     
     Inc(i);
  end;
  
  // Fallback: If disconnected loops or orphans exist (queue didn't reach everything)
  // Append remaining items that are not in sortedList?
  // This is expensive to check.
  // Simple check: Count.
  // sortedList has Header + Processed.
  // respList has Header + All.
  
  if sortedList.Count < respList.Count then
  begin
      // Add unprocessed items in their original file order
      // Iterate original list, check if present in sorted?
      // Parsing ID again is easier.
      // Optimization: Build a Set/Map of processed IDs?
      // Or just append *all* non-roots that weren't visited?
      // Since this is a fallback for broken data, simple append is OK.
      // But we don't want duplicates.
      
      // Let's just iterate original and add if not in sorted.
      // Very slow O(N^2).
      // Given typical dialogue counts (10-50), it is negligible.
      
      for i := 1 to respList.Count - 1 do
      begin
           json := respList[i];
           // Simple string comparison might fail if whitespace differs? 
           // Objects are same strings.
           if sortedList.IndexOf(json) < 0 then
              sortedList.Add(json);
      end;
  end;
  
  // Reverse the order of elements (excluding header at index 0)
  i := 1;
  j := sortedList.Count - 1;
  while i < j do
  begin
    json := sortedList[i];
    sortedList[i] := sortedList[j];
    sortedList[j] := json;
    Inc(i);
    Dec(j);
  end;

  // Copy back
  respList.Assign(sortedList);
  
  // Signup
  queue.Free;
  sortedList.Free;
  for i := 0 to pnamMap.Count - 1 do
      TObject(pnamMap.Objects[i]).Free;
  pnamMap.Free;
end;

procedure JoinList(sList: TStringList; var result: string);
var
  i: integer;
begin
  for i := 0 to sList.Count - 1 do
  begin
    result := result + sList[i];
    if i < sList.Count - 1 then result := result + ',' + #13#10
    else result := result + #13#10;
  end;
end;

function Finalize: integer;
var
  filename, filepath, jsonContent: string;
  outputFile: TStringList;
  dlg: TSaveDialog;
  i, j: integer;
  dialHeader, finalDialEntry: string;
  innerList: TStringList;
  allResps: string;
begin
  if targetFileName = '' then
  begin
    AddMessage('[ExportData] No records processed.');
    Result := 1;
    Exit;
  end;

  AddMessage('[ExportData] Finalizing...');
  AddMessage('[ExportData] Merging ' + IntToStr(dialDataMap.Count) + ' dialogue groups...');
  
  for i := 0 to dialDataMap.Count - 1 do
  begin
       if dialDataMap.Objects[i] <> nil then
       begin
            innerList := TStringList(dialDataMap.Objects[i]);
            
            // Index 0 is header. Index 1+ are responses.
            if innerList.Count > 0 then
            begin
                 dialHeader := innerList[0];
                 
                 // If Header is empty, it means we found INFOs but no DIAL record. 
                 // We should probably skip or handle gracefully.
                 if dialHeader <> '' then
                 begin
                      // Sort INFOs by PNAM chain
                      ApplyInfoSorting(innerList);

                      finalDialEntry := dialHeader + #13#10 + '  "responses": [';
                      
                      if innerList.Count > 1 then
                      begin
                           for j := 1 to innerList.Count - 1 do
                           begin
                                finalDialEntry := finalDialEntry + #13#10 + innerList[j];
                                if j < innerList.Count - 1 then
                                   finalDialEntry := finalDialEntry + ',';
                           end;
                      end;
                      
                      finalDialEntry := finalDialEntry + #13#10 + '  ]' + #13#10 + '  }';
                      dialogueList.Add(finalDialEntry);
                 end;
            end;
       end;
  end;
  
  filename := targetFileName + '_Export.json';
  
  dlg := TSaveDialog.Create(nil);
  try
    dlg.Filter := 'JSON files (*.json)|*.json|All files (*.*)|*.*';
    dlg.DefaultExt := 'json';
    dlg.FileName := filename;
    dlg.InitialDir := ProgramPath + 'Edit Scripts\';
    
    if dlg.Execute then
      filepath := dlg.FileName
    else
    begin
      AddMessage('[ExportData] Export cancelled.');
      Result := 1;
      Exit;
    end;
  finally
    dlg.Free;
  end;
  
  jsonContent := '{' + #13#10 +
    '  "target_plugin": ' + JsonString(targetFileName) + ',' + #13#10 +
    '  "dialogue_groups": [' + #13#10;
  JoinList(dialogueList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "quests": [' + #13#10;
  JoinList(questList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "items": [' + #13#10;
  JoinList(itemList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "magic": [' + #13#10;
  JoinList(magicList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "locations": [' + #13#10;
  JoinList(locationList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "cells": [' + #13#10;
  JoinList(cellList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "system": [' + #13#10;
  JoinList(systemList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "messages": [' + #13#10;
  JoinList(messageList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "load_screens": [' + #13#10;
  JoinList(loadScreenList, jsonContent);
  jsonContent := jsonContent + '  ],' + #13#10 +
    '  "npcs": {' + #13#10;
  JoinList(npcList, jsonContent);
  jsonContent := jsonContent + '  }' + #13#10 +
    '}';
  
  outputFile := TStringList.Create;
  try
    outputFile.Text := jsonContent;
    outputFile.SaveToFile(filepath);
    AddMessage('[ExportData] Export complete: ' + filename);
  finally
    outputFile.Free;
    
    // Free Map Objects
    for i := 0 to dialDataMap.Count - 1 do
        if dialDataMap.Objects[i] <> nil then
           TObject(dialDataMap.Objects[i]).Free;
    dialDataMap.Free;
    
    dialogueList.Free;
    questList.Free;
    itemList.Free;
    magicList.Free;
    locationList.Free;
    cellList.Free;
    systemList.Free;
    messageList.Free;
    loadScreenList.Free;
    npcList.Free;
    processedNPCs.Free;
    lstRecursion.Free;
  end;

  Result := 0;
end;

end.