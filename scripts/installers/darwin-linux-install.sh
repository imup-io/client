#!/bin/bash

# ======================== imUp Client Installer ======================== #
# This script installs the imUp client on MacOS and Linux systems.        #
#                                                                         #
# ----------------------------- Dependencies ---------------------------- #
# - cat                                                                   #
# - chmod (mac only, sets permissions on the launch agent plist file)     #
# - chown (mac only, sets permissions on the launch agent plist file)     #
# - curl (used to download imUp)                                          #
# - grep                                                                  #
# - launchctl (mac only)                                                  #
# - mkdir (makes the imUp install directory)                              #
# - mktemp (used to create a temp directory to store the imUp tarball)    #
# - pkill (used to kill the imUp process during uninstall)                #
# - printf                                                                #
# - sed                                                                   #
# - sudo (linux if != root user, and mac always)                          #
# - systemctl (linux only)                                                #
# - tar                                                                   #
# - uname                                                                 #
# ======================================================================= #

# exit on any error
set -e

# =================== [START] Constants and Variables =================== #
# TODO: doesn't do anything yet
install_script_version=0.2.0
build="stable"

operating_system=$(uname | tr '[:upper:]' '[:lower:]')

imup_github_releases_url="https://github.com/imup-io/client/releases"
imup_github_downloads_url="$imup_github_releases_url/download"
# this is a special URL that redirects to the latest tag
imup_github_latest_url="$imup_github_releases_url/latest"

# TODO: make sure $HOME exists or let them set a variable for this?
imup_install_dir="$HOME/imUp"

terms_of_service_error_message="
Error: You must accept the Terms of Service in order to install the imUp client. https://app.imup.io/terms
Please set the ACCEPT_THE_TOS environment variable to 'yes' to accept the Terms of Service.
"

required_dependencies=(
  "cat"
  "curl"
  "grep"
  "ls"
  "mkdir"
  "mktemp"
  "pkill"
  "printf"
  "rm"
  "sed"
  "tar"
  "uname"
)

required_dependencies_darwin=(
  "chmod"
  "chown"
  "launchctl"
  "sudo"
)

required_dependencies_linux=(
  "systemctl"
)

declare -a options=()
# ==================== [END] Constants and Variables ==================== #

# ========================== [START] Functions ========================== #
# use sudo if the user isn't root
sudo() {
  if [ "$EUID" -ne 0 ]
  then
    command sudo "$@"
  else
    "$@"
  fi
}

uninstall_client_darwin() {
  # terminate the imUp process
  sudo pkill -f "imUp" || true

  # Remove the imUp.plist file
  sudo launchctl unload /Library/LaunchAgents/imUp.plist || true
  sudo rm /Library/LaunchAgents/imUp.plist || true

  # remove the imUp files
  sudo rm -rf "$imup_install_dir"

  exit 0
}

uninstall_client_linux() {
  # terminate the imUp process
  sudo pkill -f "imUp" || true

  # Remove the imUp.service file
  sudo systemctl stop imUp.service
  sudo systemctl disable imUp.service
  sudo rm -f /etc/systemd/system/imUp.service
  sudo systemctl daemon-reload

  # remove the imUp files
  sudo rm -rf "$imup_install_dir" || true

  exit 0
}

# exit with error message if the user doesn't accept the ToS
check_user_accepted_tos() {
  if [ "$ACCEPT_THE_TOS" != "yes" ]
  then
    echo "$terms_of_service_error_message"
    exit 1
  fi
}

# if any of the required dependencies are not installed, prints
# an error message with the missing dependencies and exits
check_required_dependencies_are_installed() {
  declare -a missing_dependencies=()

  for dependency in "${required_dependencies[@]}"
  do
    if ! command -v "$dependency" &> /dev/null
    then
      missing_dependencies+=("$dependency")
    fi
  done

  # if on mac, check for mac specific dependencies
  if [ "$operating_system" == "darwin" ]
  then
    for dependency in "${required_dependencies_darwin[@]}"
    do
      if ! command -v "$dependency" &> /dev/null
      then
        missing_dependencies+=("$dependency")
      fi
    done
  fi

  # if on linux, check for linux specific dependencies
  if [ "$operating_system" == "linux" ]
  then
    for dependency in "${required_dependencies_linux[@]}"
    do
      if ! command -v "$dependency" &> /dev/null
      then
        missing_dependencies+=("$dependency")
      fi
    done
  fi

  # if there are any missing dependencies, print an error message and exit
  if [ ${#missing_dependencies[@]} -ne 0 ]
  then
    echo "Error: The following dependencies are required to install the imUp client:"
    for dependency in "${missing_dependencies[@]}"
    do
      echo "- $dependency"
    done
    exit 1
  fi
}

# exits if required environment variables are not set
check_required_environment_variables() {
  if [ -z "$API_KEY" ] && [ -z "$EMAIL" ]; then
    echo "Error: One or both of the 'API_KEY' and 'EMAIL' environment variables need to be set."
    exit 1
  fi
}

# read environment variables and build an array
# of options to pass to the client
create_options_array() {
  if [ -n "$ALLOWLISTED_IPS" ]; then
    options+=("--allowlisted-ips='$ALLOWLISTED_IPS'")
  fi
  if [ -n "$API_KEY" ]; then
    options+=("--key='$API_KEY'")
  fi
  if [ -n "$BLOCKLISTED_IPS" ]; then
    options+=("--blocklisted-ips='$BLOCKLISTED_IPS'")
  fi
  if [ -n "$EMAIL" ]; then
    options+=("--email='$EMAIL'")
  fi
  if [ -n "$GROUP_ID" ]; then
    options+=("--group-id='$GROUP_ID'")
  fi
  if [ -n "$HOST_ID" ]; then
    options+=("--host-id='$HOST_ID'")
  fi
  if [ -n "$IMUP_ADDRESS" ]; then
    options+=("--api-post-connection-data='$IMUP_ADDRESS'")
  fi
  if [ -n "$IMUP_ADDRESS_SPEEDTEST" ]; then
    options+=("--api-post-speed-test-data='$IMUP_ADDRESS_SPEEDTEST'")
  fi
  if [ "$NO_SPEED_TEST" == "true" ]; then
    options+=("--no-speed-test")
  fi
  if [ -n "$PING_ADDRESS" ]; then
    options+=("--ping-addresses-external='$PING_ADDRESS'")
  fi
  if [ -n "$PING_ADDRESS_INTERNAL" ]; then
    options+=("--ping-address-internal='$PING_ADDRESS_INTERNAL'")
  fi
  if [ -n "$PING_DELAY" ]; then
    options+=("--ping-delay='$PING_DELAY'")
  fi
  if [ -n "$PING_ENABLED" ]; then
    options+=("--ping='$PING_ENABLED'")
  fi
  if [ -n "$PING_INTERVAL" ]; then
    options+=("--ping-interval='$PING_INTERVAL'")
  fi
  if [ -n "$PING_REQUESTS" ]; then
    options+=("--ping-requests='$PING_REQUESTS'")
  fi
}

# pretty prints the options array
print_options_array() {
  echo "Options:"
  for option in "${options[@]}"
  do
    echo "- $option"
  done
  echo ""
}

log_msg() {
  msg=$1
  printf "%s\n" "$msg"
}

log_msg_blue() {
  msg=$1
  printf "\033[34m%s\033[0m\n" "$msg"
}

log_msg_red() {
  msg=$1
  printf "\033[31m%s\033[0m\n" "$msg"
}

# sets architecture variable based on uname -m
get_arch() {
  set +e

  uname -m | grep -q x86_64
  is_x86="$?"

  uname -a | grep -q arm64
  is_arm64="$?"

  uname -m | grep -q 'i.*86'
  is_32bit="$?"

  architecture="$(uname -m)"

  set -e

  # account for running under rosetta
  if [ "$is_x86" -eq 0 ]; then
  	if [ "$is_arm64" -eq 0 ]; then
  		architecture='arm64'
  	else
  		architecture='amd64'
  	fi
  elif [ "$is_32bit" -eq 0 ]; then
  	architecture='x32'
  elif [ "$architecture" = "arm64" ]; then
    architecture='arm64'
  elif [ "$architecture" = "aarch64" ]; then
  	architecture='arm64'
  else
    log_msg_red "Architecture '$architecture' not supported"
  	exit 1
  fi
}

# creates a temporary directory and sets the imup_tarball_filepath variable
get_temp_tarball_filepath() {
  # create a temporary directory to store the imUp tarball
  imup_temp_dir=$(mktemp -d -t 'imup-installer.XXXXXX')
  imup_tarball_filepath="$imup_temp_dir/imUp-$operating_system-$architecture.tar.gz"
}

# gets the latest imUp release version and builds the download URL
get_download_url() {
  # follow the redirect and get the URL
  imup_github_latest_tag_url=$(curl -s -L -o /dev/null -w '%{url_effective}' "$imup_github_latest_url")
  # get the version number from the URL
  imup_github_latest_release_version=$(echo "$imup_github_latest_tag_url" | sed -e 's/.*\/v//g')

  imup_github_asset_name="client_"$imup_github_latest_release_version"_"$operating_system"_$architecture.tar.gz"

  download_url="$imup_github_downloads_url/v$imup_github_latest_release_version/$imup_github_asset_name"
}

download_and_install_imup() {
  # Download and install the client
  log_msg_blue "Downloading imUp client from $download_url"
  curl --fail --progress-bar -L $download_url -o $imup_tarball_filepath

  log_msg_blue "Extracting client to $imup_install_dir"

  # make clean install dir
  rm -rf $imup_install_dir
  mkdir -p $imup_install_dir

  # extract tarball directly to install dir
  tar -xzf $imup_tarball_filepath -C $imup_install_dir
  rm -f $imup_tarball_filepath

  echo ""
  echo "Installed contents at $imup_install_dir:"
  ls -la $imup_install_dir
}

install_macos_launch_agent() {
  # Set the launch agent properties
  tmpPLIST_PATH="./imUp.plist"
  PLIST_PATH="/Library/LaunchAgents/imUp.plist"
  LABEL="io.imUp"
  RUN_AT_LOAD="true"
  USER_ID="0"

  # build the ProgramArguments array for the .plist file
  program_arguments=""
  program_arguments+="\n        <string>$imup_install_dir/imUp</string>"
  for option in "${options[@]}"; do
      if [ ! -z "$option" ]; then
          program_arguments+="\n        <string>$option</string>"
      fi
  done
  program_arguments+="\n"

  # create the .plist file
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

  # set the correct permissions on the launch agent plist file
  sudo cp $tmpPLIST_PATH $PLIST_PATH
  sudo chown root:wheel "$PLIST_PATH"
  sudo chmod 644 "$PLIST_PATH"

  # load the launch agent
  echo ""
  echo "Loading launch agent..."
  sudo launchctl load "$PLIST_PATH"

  echo "Launch agent installed at $PLIST_PATH"
}

install_linux_systemd_service() {
  # create the systemd service file
  cat > /etc/systemd/system/imUp.service <<EOF
[Unit]
Description=imUp service

[Service]
ExecStart=$imup_install_dir/imup ${options[@]}

[Install]
WantedBy=multi-user.target
EOF

  # Enable and start the systemd service
  systemctl enable imUp.service
  systemctl start imUp.service
}
# =========================== [END] Functions =========================== #

# ========================= [START] Main Script ========================= #

# if the UNINSTALL env var is set to "true", uninstall the client and exit
if [ "$UNINSTALL" ]
then
  if [ "$operating_system" == "darwin" ]
  then
    uninstall_client_darwin
  elif [ "$operating_system" == "linux" ]
  then
    uninstall_client_linux
  fi
fi

check_user_accepted_tos

check_required_dependencies_are_installed

check_required_environment_variables

get_temp_tarball_filepath

get_arch

get_download_url

create_options_array

print_options_array

on_error() {
  printf "\033[31m$ERROR_MESSAGE
It looks like you hit an issue when trying to install the imUp Client. Please reach out to support@imup.io for help.\n\033[0m\n"
}
trap on_error ERR

download_and_install_imup

# MacOS launch agent install
if [ "$operating_system" == "darwin" ]
then
  install_macos_launch_agent
fi

# linux systemd install
if [ "$operating_system" == "linux" ]
then
  install_linux_systemd_service
fi
