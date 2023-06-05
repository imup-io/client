$ErrorActionPreference = "Stop"

$install_script_version = "0.2.0"
$file = "$env:TEMP\imUp-windows-amd64.zip"
$build = "stable"
$base_url = "https://github.com/imup-io/client/releases/download/"
$release_version = "v0.17.0"
$remote_file = "client_0.17.0_windows_amd64.zip"
$url = "$base_url/$release_version/$remote_file"
$install_dir = "$env:UserProfile\imUp"

$key = $env:key
$email = $env:email
$host_id = $env:host_id
$group_id = $env:group_id
$SpeedTestEnabled = $env:SpeedTestEnabled
$allowlisted_ips = $env:allowlisted_ips
$blocklisted_ips = $env:blocklisted_ips
$ping = $env:ping
$ping_addresses_external = $env:ping_addresses_external
$ping_address_internal = $env:ping_address_internal
$ping_delay = $env:ping_delay
$ping_interval = $env:ping_interval
$ping_requests = $env:ping_requests

if not "%email%" == "" set "serviceArguments=!serviceArguments! --email=!email!"
if not "%key%" == "" set "serviceArguments=!serviceArguments! --key=!key!"
if not "%host_id%" == "" set "serviceArguments=!serviceArguments! --host-id=!host_id!"
if not "%group_id%" == "" set "serviceArguments=!serviceArguments! --group-id=!group_id!"
if not "%SpeedTestEnabled%" == "" set "serviceArguments=!serviceArguments! --SpeedTestEnabled=!SpeedTestEnabled!"
if not "%allowlisted_ips%" == "" set "serviceArguments=!serviceArguments! --allowlisted-ips=!allowlisted_ips!"
if not "%blocklisted_ips%" == "" set "serviceArguments=!serviceArguments! --blocklisted-ips=!blocklisted_ips!"
if not "%ping%" == "" set "serviceArguments=!serviceArguments! --ping=!ping!"
if not "%ping_addresses_external%" == "" set "serviceArguments=!serviceArguments! --ping-addresses-external=!ping_addresses_external!"
if not "%ping_address_internal%" == "" set "serviceArguments=!serviceArguments! --ping-address-internal=!ping_address_internal!"
if not "%ping_delay%" == "" set "serviceArguments=!serviceArguments! --ping-delay=!ping_delay!"
if not "%ping_interval%" == "" set "serviceArguments=!serviceArguments! --ping-interval=!ping_interval!"
if not "%ping_requests%" == "" set "serviceArguments=!serviceArguments! --ping-requests=!ping_requests!"


if (-not $email) {
    Write-Host "Email not provided. Please set the email address in the environment variable `email`." -ForegroundColor Red
    exit 1
}


# Download and install the client
Write-Host "Downloading imUp client"
Remove-Item -Path $file -Force -ErrorAction SilentlyContinue
Invoke-WebRequest -Uri $url -OutFile $file -ErrorAction Stop

Write-Host "Extracting client"
Remove-Item -Path $install_dir -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $install_dir | Out-Null
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory($file, $install_dir)
Get-ChildItem -Path $install_dir
Remove-Item -Path $file -Force -ErrorAction SilentlyContinue

# Create the windows service that runs the extracted imUp binary
$serviceName = "imUp13"
$serviceExePath = Join-Path $install_dir "imUp.exe"
$serviceDescription = "imUp.io is an internet performance measurement tool"
$serviceCredential = $null
$serviceArguments = "--email=$email -log-to-file=true --key=$key --host-id=$host-id"
$serviceDisplayName = $serviceName
$serviceStartMode = "Automatic"
$type = "Own"

# Create the service
$serviceArgs = @{
	Name = $serviceName
    DisplayName = $serviceName
    Description = $serviceDescription
    BinaryPathName = "$serviceExePath $serviceArguments"
    start = $serviceStartMode
}

Write-Host @serviceArgs
New-Service @serviceArgs



Write-Host "imUp service successfully installed"