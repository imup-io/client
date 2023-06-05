#!/bin/sh

set -e
install_script_version=0.2.0
file="/tmp/imUp-darwin-amd64.tar.gz"
build="stable"
base_url="https://github.com/imup-io/client/releases/download/"
release_version="v0.14.0"
remote_file="client_0.14.0_darwin_amd64.tar.gz"
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

params = [
  KEY="$KEY"
  GROUP="$GROUP"
  EMAIL="$EMAIL"
  UNINSTALL="$UNINSTALL"
  ALLOWLISTED_IPS="$ALLOWLISTED_IPS"
  BLOCKLISTED_IPS="$BLOCKLISTED_IPS"
  PING="$PING"
  PING_ADDRESS="$PING_ADDRESS"
  PING_ADDRESS_INTERNAL="$PING_ADDRESS_INTERNAL"
  PING_DELAY="$PING_DELAY"
  PING_INTERVAL="$PING_INTERVAL"
  PING_REQUESTS="$PING_REQUESTS"
]


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
ls -la $install_dir
rm -f $file

# Create the systemd service file
cat > $SYSTEMD_SERVICE_FILE
<<EOF
[Unit]
Description=$BINARY service

[Service]
ExecStart=$install_dir/imUp
$PARAMETERS

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the systemd service
service

systemctl enable $SYSTEMD_SERVICE_NAME
systemctl start $SYSTEMD_SERVICE_NAME