# Installation

## Download the latest release from GitHub
https://github.com/Keyfactor/kfutil/releases

### Via Github CLI

#### Linux/macOS
The following attempts to
```bash
gh auth login
export REPO_NAMESPACE=Keyfactor
export REPO_NAME=kfutil
export REPO_PATH="${REPO_NAMESPACE}/${REPO_NAME}"
export TAG=$(gh release list --repo $REPO_PATH --limit 1 | cut -f 1) #Get the latest release
export RELEASE_DIR="${REPONAME}-${TAG}"
mkdir -p $RELEASE_DIR
cd $RELEASE_DIR && gh release download $TAG --repo $REPO_PATH --clobber
```

#### Windows Powershell
```powershell
gh auth login
$env:REPO_NAMESPACE="Keyfactor"
$env:REPO_NAME="kfutil"
$env:REPO_PATH="${REPO_NAMESPACE}/${REPO_NAME}"
$env:TAG=$(gh release list --repo $REPO_PATH --limit 1 | Select-String -Pattern "v\d+\.\d+\.\d+" -AllMatches | ForEach-Object { $_.Matches.Value }) #Get the latest release
$env:RELEASE_DIR="${REPONAME}-${TAG}"
mkdir -p $RELEASE_DIR
cd $RELEASE_DIR && gh release download $TAG --repo $REPO_PATH --clobber
```
## Install kfutil
Once you've downloaded a release version you can install kfutil by running the following command:  

### Linux
```bash
unzip kfutil_<version>_linux_amd64.zip
sudo mv kfutil_<version> /usr/local/bin/kfutil
```

### macOS
```bash
unzip kfutil_<version>_darwin_amd64.zip
sudo mv kfutil_<version> /usr/local/bin/kfutil
```

### Windows Powershell
```powershell
Expand-Archive kfutil_<version>_windows_amd64.zip
Move-Item .\kfutil_<version>.exe C:\Windows\System32\kfutil.exe
```

## Configure kfutil

## Linux/macOS
```bash
export KEYFACTOR_HOSTNAME=<mykeyfactorhost.mydomain.com>
export KEYFACTOR_USERNAME=<myusername> # Do not include domain
export KEYFACTOR_PASSWORD=<mypassword>
export KEYFACTOR_DOMAIN=<mykeyfactordomain>
```

## Windows Powershell
```powershell
$env:KEYFACTOR_HOSTNAME="<mykeyfactorhost.mydomain.com>"
$env:KEYFACTOR_USERNAME="<myusername>" # Do not include domain
$env:KEYFACTOR_PASSWORD="<mypassword>"
$env:KEYFACTOR_DOMAIN="<mykeyfactordomain>"
```
## Run kfutil

### Linux/macOS/Windows Powershell
```bash
kfutil
```
