Unicode true

####
## User mode installation settings
!define REQUEST_EXECUTION_LEVEL "user"
####

!include "wails_tools.nsh"

# Version information
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support
ManifestDPIAware true

!include "MUI.nsh"

Var RunAppPath  # Variable to store application path

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

# Configure finish page with run checkbox
!define MUI_FINISHPAGE_RUN $RunAppPath
!define MUI_FINISHPAGE_RUN_TEXT "Start ${INFO_PRODUCTNAME} on exit"
!define MUI_FINISHPAGE_RUN_CHECKED

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$LOCALAPPDATA\Programs\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
ShowInstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext

    # Install WebView2 Runtime if needed
    !insertmacro wails.webview2runtime

    # Set output path to installation directory
    SetOutPath $INSTDIR

    # Install application files
    !insertmacro wails.files

    # Store application path for finish page
    StrCpy $RunAppPath "$INSTDIR\${PRODUCT_EXECUTABLE}"

    # Create user-specific shortcuts
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    # Register file associations and protocols
    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    # Write uninstaller
    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    # Remove application data
    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"

    # Remove installation directory
    RMDir /r $INSTDIR

    # Remove shortcuts
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    # Unregister file associations and protocols
    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    # Remove uninstaller
    !insertmacro wails.deleteUninstaller
SectionEnd
