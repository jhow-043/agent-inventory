; ============================================================================
; Inventory Agent - Inno Setup Installer Script
; ============================================================================
; Builds a Windows installer that:
;   1. Installs inventory-agent.exe to {app}
;   2. Asks for Server URL during setup
;   3. Generates config.json with embedded enrollment key + user URL
;   4. Registers and starts the InventoryAgent Windows Service
;   5. On uninstall, stops/removes the service and deletes files
;
; Build with:  ISCC /DEnrollmentKey="your-key-here" setup.iss
; ============================================================================

#define MyAppName      "Inventory Agent"
#define MyAppVersion   "0.1.0"
#define MyAppPublisher  "IT Inventory"
#define MyAppExeName   "inventory-agent.exe"

; Enrollment key must be provided at compile time via /DEnrollmentKey="..."
#ifndef EnrollmentKey
  #error "You must pass /DEnrollmentKey=\"your-key\" to ISCC"
#endif

[Setup]
AppId={{A1B2C3D4-E5F6-7890-ABCD-EF1234567890}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\InventoryAgent
DisableDirPage=yes
DisableProgramGroupPage=yes
OutputDir=output
OutputBaseFilename=InventoryAgentSetup-{#MyAppVersion}
Compression=lzma2/ultra64
SolidCompression=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
UninstallDisplayIcon={app}\{#MyAppExeName}
UninstallDisplayName={#MyAppName}
; SetupIconFile=..\assets\icon.ico  ; Uncomment when you have a custom icon
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "brazilianportuguese"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"

[Files]
Source: "..\build\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

; ============================================================================
; Custom Input Pages — Server URL and Enrollment Key
; ============================================================================
[Code]

var
  ConfigPage: TInputQueryWizardPage;

procedure InitializeWizard();
begin
  ConfigPage := CreateInputQueryPage(
    wpSelectDir,
    'Server Configuration',
    'Enter the inventory server address.',
    'The enrollment key is already embedded in this installer.'
  );
  ConfigPage.Add('Server URL:', False);

  { Default }
  ConfigPage.Values[0] := 'https://inventory.example.com';
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = ConfigPage.ID then
  begin
    if Trim(ConfigPage.Values[0]) = '' then
    begin
      MsgBox('Server URL cannot be empty.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
  end;
end;

{ Write config.json after files are copied }
procedure CurStepChanged(CurStep: TSetupStep);
var
  ConfigFile: String;
  Lines: TStringList;
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    ConfigFile := ExpandConstant('{app}\config.json');
    Lines := TStringList.Create;
    try
      Lines.Add('{');
      Lines.Add('  "server_url": "' + Trim(ConfigPage.Values[0]) + '",');
      Lines.Add('  "enrollment_key": "{#EnrollmentKey}",');
      Lines.Add('  "interval_hours": 1,');
      Lines.Add('  "data_dir": "",');
      Lines.Add('  "log_level": "info",');
      Lines.Add('  "insecure_skip_verify": false');
      Lines.Add('}');
      Lines.SaveToFile(ConfigFile);
    finally
      Lines.Free;
    end;

    { Install and start the Windows service }
    Exec(
      ExpandConstant('{app}\{#MyAppExeName}'),
      'install',
      ExpandConstant('{app}'),
      SW_HIDE, ewWaitUntilTerminated, ResultCode
    );
    Exec(
      ExpandConstant('{app}\{#MyAppExeName}'),
      'start',
      ExpandConstant('{app}'),
      SW_HIDE, ewWaitUntilTerminated, ResultCode
    );
  end;
end;

{ Stop and uninstall the service before files are removed }
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  ResultCode: Integer;
begin
  if CurUninstallStep = usUninstall then
  begin
    Exec(
      ExpandConstant('{app}\{#MyAppExeName}'),
      'stop',
      ExpandConstant('{app}'),
      SW_HIDE, ewWaitUntilTerminated, ResultCode
    );
    Sleep(2000);
    Exec(
      ExpandConstant('{app}\{#MyAppExeName}'),
      'uninstall',
      ExpandConstant('{app}'),
      SW_HIDE, ewWaitUntilTerminated, ResultCode
    );
    Sleep(1000);
    { Remove config and data files }
    DeleteFile(ExpandConstant('{app}\config.json'));
    DelTree(ExpandConstant('{app}\data'), True, True, True);
    DeleteFile(ExpandConstant('{app}\token.json'));
  end;
end;
