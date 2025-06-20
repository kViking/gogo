#define Version GetEnv('INNO_VERSION')

[Setup]
AppName=GoGoGadget
AppVersion={#Version}
VersionInfoVersion={#Version}
DefaultDirName={userappdata}\GoGoGadget
DefaultGroupName=GoGoGadget
OutputBaseFilename=GoGoGadget-Installer-{#Version}
Compression=lzma
SolidCompression=yes
ChangesEnvironment=yes
PrivilegesRequired=lowest
; Add these lines to minimize user interaction
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
DisableFinishedPage=yes
DisableWelcomePage=no

[Files]
; Main executable
Source: "GoGoGadget.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "GoGoGadget.ps1"; DestDir: "{app}"; Flags: ignoreversion
Source: "user_scripts.json"; DestDir: "{userappdata}\GoGoGadget"; Flags: onlyifdoesntexist
Source: "settings.json"; DestDir: "{userappdata}\GoGoGadget"; Flags: onlyifdoesntexist
; Installation helper scripts
Source: "install-profile.ps1"; DestDir: "{tmp}"; Flags: deleteafterinstall
Source: "uninstall-profile.ps1"; DestDir: "{tmp}"; Flags: deleteafterinstall

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"

[Run]
Filename: "powershell.exe"; \
  Parameters: "-NoProfile -ExecutionPolicy Bypass -File '{tmp}\install-profile.ps1' -GadgetPath='{app}\GoGoGadget.ps1'"; \
  Flags: runhidden; \
  StatusMsg: "Configuring PowerShell profile for GoGoGadget completions..."

[UninstallRun]
; No warning dialog, just clean up profile
Filename: "powershell.exe"; \
  Parameters: "-NoProfile -ExecutionPolicy Bypass -File '{tmp}\uninstall-profile.ps1' -GadgetPath='{app}\GoGoGadget.ps1'"; \
  Flags: runhidden; \
  StatusMsg: "Cleaning up GoGoGadget completions from PowerShell profile..."; \
  RunOnceId: GoGoGadgetUninstallCleanup

[UninstallDelete]
Type: filesandordirs; Name: "{userappdata}\GoGoGadget"

[Code]
function StringReplace(const S, OldPattern, NewPattern: String): String;
var
  SearchStr, Patt, NewStr: String;
  Offset: Integer;
begin
  SearchStr := S;
  Patt := OldPattern;
  NewStr := '';
  while Length(SearchStr) > 0 do
  begin
    Offset := Pos(Patt, SearchStr);
    if Offset = 0 then
    begin
      NewStr := NewStr + SearchStr;
      Break;
    end;
    NewStr := NewStr + Copy(SearchStr, 1, Offset - 1) + NewPattern;
    SearchStr := Copy(SearchStr, Offset + Length(Patt), MaxInt);
  end;
  Result := NewStr;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  Path, AppPath, NewPath: string;
begin
  if CurUninstallStep = usUninstall then
  begin
    AppPath := ExpandConstant('{app}');
    if RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Path) then
    begin
      // Remove our app path from the PATH variable
      NewPath := StringReplace(Path, ';' + AppPath, '');
      NewPath := StringReplace(NewPath, AppPath + ';', '');
      NewPath := StringReplace(NewPath, AppPath, '');
      RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', NewPath);
    end;
  end;
end;
