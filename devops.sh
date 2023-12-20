# /bin/bash

KATZENMINT=0
MIX=0
PROVIDER=0
ERROR=0
CONFIGDIR=$(pwd)
BINARYDIR=$(pwd)

while [ $# -gt 0 ]; do
	case "$1" in
		--katzenmint)
			KATZENMINT=1
			;;
		--mix)
			MIX=1
			;;
		--provider)
			PROVIDER=1
			;;
    --configdir)
      CONFIGDIR="$2"
      shift
      ;;
    --binarydir)
      BINARYDIR="$2"
      shift
      ;;
		*)
			echo "Illegal option $1"
			ERROR=1
			;;
	esac
	shift $(( $# > 0 ? 1 : 0 ))
done

usage () {
	cat <<EOF

Usage:
  sh devops.sh [options]
  example: sh devops --katzenmint

Available options:
  --katzenmint update katzenmint
  --mix update mix
  --provider update provider
  --configdir config directory
  --binarydir binary directory
EOF
	exit 1
}

stopRemoteService () {
  echo "stop $1..."
  ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l $1 testnet/remote/ansible/stop.yml
  if [ $? -ne 0 ]; then
    echo "failed to stop $1"
    exit 1
  fi
}

main () {
  if [ ! -d "$CONFIGDIR" ]; then
    echo "config directory not existed"
    exit 1
  fi
  if [ ! -d "$BINARYDIR" ]; then
    echo "binary directory not existed"
    exit 1
  fi
  if [ $KATZENMINT = 1 ]; then
    echo "update katzenmints..."
    
    stopRemoteService "mixnet"
    stopRemoteService "providernet"
    stopRemoteService "sentrynet"

    echo "upload binary files"
    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l sentrynet testnet/remote/ansible/config.yml -e CONFIGDIR=$CONFIGDIR -e BINARY=$BINARYDIR/katzenmint
    if [ $? -ne 0 ]; then
      echo "failed to upload"
    fi
    
    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l sentrynet testnet/remote/ansible/status.yml
  fi

  if [ $MIX = 1 ]; then
    if [ $KATZENMINT = 0 ]; then
      stopRemoteService "mixnet"
    fi

    echo "upload binary files"
    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l mixnet testnet/remote/ansible/config.yml -e CONFIGDIR=$CONFIGDIR -e BINARY=$BINARYDIR/meson-server
    if [ $? -ne 0 ]; then
      echo "failed to upload"
    fi

    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l mixnet testnet/remote/ansible/status.yml
  fi

  if [ $PROVIDER = 1 ]; then
    if [ $KATZENMINT = 0 ]; then
      stopRemoteService "providernet"
    fi

    echo "upload binary files"
    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l providernet testnet/remote/ansible/config.yml -e CONFIGDIR=$CONFIGDIR -e BINARY=$BINARYDIR
    if [ $? -ne 0 ]; then
      echo "failed to upload"
    fi
    
    ansible-playbook -i testnet/remote/ansible/inventory/digital_ocean.py -l providernet testnet/remote/ansible/status.yml
  fi
}

if [ $KATZENMINT = 0 ] && [ $MIX = 0 ] && [ $PROVIDER = 0 ]; then
	usage
fi

if [ $ERROR = 1 ]; then
	usage
fi



main