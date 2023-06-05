#!/bin/sh

set -e
install_script_version=0.2.0
file="/tmp/imUp-darwin-amd64.tar.gz"
build="stable"
base_url="https://github.com/imup-io/client/releases/download/"
release_version="v0.18.0"
remote_file="client_0.18.0_darwin_amd64.tar.gz"
DOWNLOAD_URL="$base_url/$release_version/$remote_file"
install_dir="$HOME/imUp"

# Uninstall the client if the uninstall flag is set and then exit.
if [ "$UNINSTALL" ]; then
  # Terminate the imUp process
  sudo pkill -f "imUp" || true

  # Remove the imUp.plist file
  sudo launchctl unload /Library/LaunchAgents/imUp.plist || true
  sudo rm /Library/LaunchAgents/imUp.plist || true

  # Remove the imUp binary
  sudo rm -rf "$install_dir" || true
  exit 0
fi

# Read the variables from the environment and catch errors
if [ -z "$KEY" ] && [ -z "$EMAIL" ]; then
  echo "Error: Either KEY or EMAIL are required."
  exit 1
fi

declare -a options=()

# Append each non-null variable to the options array
if [ -n "$KEY" ]; then
  options+=("--key='$KEY'")
fi
if [ -n "$GROUP" ]; then
  options+=("--group-id='$GROUP'")
fi
if [ -n "$EMAIL" ]; then
  options+=("--email='$EMAIL'")
fi
if [ -n "$ID" ]; then
  options+=("--host-id='$ID'")
fi
if [ "$SPEEDTEST_ENABLED" == "false" ]; then
  options+=("--no-speed-test")
fi
if [ -n "$ALLOWLISTED_IPS" ]; then
  options+=("--allowlisted-ips='$ALLOWLISTED_IPS'")
fi
if [ -n "$BLOCKLISTED_IPS" ]; then
  options+=("--blocklisted-ips='$BLOCKLISTED_IPS'")
fi
if [ -n "$PING" ]; then
  options+=("--ping='$PING'")
fi
if [ -n "$PING_ADDRESSES_EXTERNAL" ]; then
  options+=("--ping-addresses-external='$PING_ADDRESSES_EXTERNAL'")
fi
if [ -n "$PING_ADDRESS_INTERNAL" ]; then
  options+=("--ping-address-internal='$PING_ADDRESS_INTERNAL'")
fi
if [ -n "$PING_DELAY" ]; then
  options+=("--ping-delay='$PING_DELAY'")
fi
if [ -n "$PING_INTERVAL" ]; then
  options+=("--ping-interval='$PING_INTERVAL'")
fi
if [ -n "$PING_REQUESTS" ]; then
  options+=("--ping-requests='$PING_REQUESTS'")
fi
if [ -n "$API_POST_SPEED_TEST_DATA"]; then
  options+=("--api-post-speed-test-data='$API_POST_SPEED_TEST_DATA'")
fi
if [ -n "$API_POST_PING_DATA"]; then
  options+=("--api-post-connection-data='$API_POST_CONNECTION_DATA'")
fi

echo "Environment: "
echo "${options[@]}"

# mac os arch detection
get_arch() {
  set +e

  uname -m | grep -q x86_64
  is_x86="$?"

  uname -a | grep -q arm64
  is_arm64="$?"

  uname -m | grep -q 'i.*86'
  is_32bit="$?"

  arch="$(uname -m)"

  set -e

  # account for running under rosetta
  if [ "$is_x86" -eq 0 ]; then
  	if [ "$is_arm64" -eq 0 ]; then
  		echo 'a64'
  	else
  		echo 'x64'
  	fi
  elif [ "$is_32bit" -eq 0 ]; then
  	echo 'x32'
  elif [ "$arch" = "arm64" ]; then
    echo 'a64'
  elif [ "$arch" = "aarch64" ]; then
  	echo 'a64'
  else
    printf "\033]31mArchitecture %s not supported\033[0m\n" "$arch"
  	exit 1
  fi
}

header() {
  msg=$1
  printf "\033[34m%s\033[0m\n" "$msg"
}

log_msg() {
  msg=$1
  printf "%s\n" "$msg"
}

check_rosetta() {
  set +e
  rosetta_installed=$(/usr/bin/pgrep oahd)
  set -e

  if [ ! $rosetta_installed ]; then
    log_msg ""
    header "M1 Support: "
    log_msg "When macOS tries to run an app that is not built for Apple silicon, macOS will prompt to "
    log_msg "install Rosetta 2 to translate the app to Apple silicon automatically. In Terminal, there "
    log_msg "is no automatic detection for missing Rosetta to run older architecture command line tools. "
    log_msg ""
    log_msg "Ensure Rosetta 2 is installed before installing the Level agent. "
    log_msg ""
    log_msg "    command:    softwareupdate --install-rosetta"
    log_msg ""
    exit 1
  fi
}

arch=$(get_arch)
# m1: check if rosetta is installed
if [ "$arch" = "a64" ]; then
  check_rosetta
fi

on_error() {
  printf "\033[31m$ERROR_MESSAGE
It looks like you hit an issue when trying to install the imUp Client. Please reach out to support@imup.io for help.\n\033[0m\n"
}
trap on_error ERR    # NOTE: trapping ERR is undefined in posix sh


# Download and install the client
header "Downloading imUp client"
rm -rf $file
curl --fail --progress-bar -L $DOWNLOAD_URL -o $file

header "Extracting client"
rm -rf $install_dir
mkdir -p $install_dir
tar -xzf $file -C $install_dir
echo ""
echo "Downloaded contents:"
ls -la $install_dir
rm -f $file


# Set the launch agent properties
tmpPLIST_PATH="./imUp.plist"
PLIST_PATH="/Library/LaunchAgents/imUp.plist"
LABEL="io.imUp"
RUN_AT_LOAD="true"
USER_ID="0"

# Build the ProgramArguments array for the .plist file
program_arguments=""
program_arguments+="\n        <string>$install_dir/imUp</string>"
for option in "${options[@]}"; do
    if [ ! -z "$option" ]; then
        program_arguments+="\n        <string>$option</string>"
    fi
done
program_arguments+="\n"

# Create the .plist file
echo "Heads up: This next step requires SUDO permissions"
echo ""
(
printf "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
printf "<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n"
printf "<plist version=\"1.0\">\n"
printf "<dict>\n"
printf "    <key>Label</key>\n"
printf "    <string>io.imUp</string>\n"
printf "    <key>ProgramArguments</key>\n"
printf "    <array>"
printf "    $program_arguments"
printf "    </array>\n"
printf "    <key>RunAtLoad</key>\n"
printf "    <true/>\n"
printf "</dict>\n"
printf "</plist>\n" 
) > "$tmpPLIST_PATH"

# Set the correct permissions on the launch agent plist file
sudo cp $tmpPLIST_PATH $PLIST_PATH
sudo chown root:wheel "$PLIST_PATH"
sudo chmod 644 "$PLIST_PATH"

# Load the launch agent
echo ""
echo "Loading launch agent..."
sudo launchctl load "$PLIST_PATH"

echo "Launch agent installed at $PLIST_PATH"