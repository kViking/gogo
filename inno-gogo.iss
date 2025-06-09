#define Version GetEnv('INNO_VERSION')

[Setup]
AppName=GoGoGadget
AppVersion={#Version}
VersionInfoVersion={#Version}
DefaultDirName={autopf}\GoGoGadget
DefaultGroupName=GoGoGadget
OutputBaseFilename=GoGoGadget-Installer-{#Version}
Compression=lzma
SolidCompression=yes
ChangesEnvironment=yes
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
; Warn the user that their saved scripts will be deleted
Filename: "powershell.exe"; \
  Parameters: '-NoProfile -Command "Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show(\'Uninstalling GoGoGadget will delete your saved scripts and settings in %APPDATA%\\GoGoGadget. If you want to keep them, make a backup now!\',\'GoGoGadget Uninstall Warning\',[System.Windows.MessageBoxButton]::OK,[System.Windows.MessageBoxImage]::Warning)"'; \
  Flags: runhidden; \
  StatusMsg: "Warning about deletion of saved scripts..."
Filename: "powershell.exe"; \
  Parameters: "-NoProfile -ExecutionPolicy Bypass -File '{tmp}\uninstall-profile.ps1' -GadgetPath='{app}\GoGoGadget.ps1'"; \
  Flags: runhidden; \
  StatusMsg: "Cleaning up GoGoGadget completions from PowerShell profile..."

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
