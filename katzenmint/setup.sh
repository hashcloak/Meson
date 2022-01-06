# /bin/sh

install_tendermint() {
    echo "Install tendermint......"
    git clone https://github.com/tendermint/tendermint.git
    cd tendermint
    git checkout v0.34.6
    make install
    cd ..
    rm -rf tendermint
    echo "Tendermint had been installed......"
}

if ! command -v tendermint &> /dev/null
then
    # echo "No tendermint in PATH, install tendermint (y/n)?"
    read -p "No tendermint in PATH, install tendermint (y/n)?"
    if [ $REPLY = "y" ]; then
        # check whether git and make installed
        if command -v git &> /dev/null && command -v make &> /dev/null; then
            install_tendermint
        fi
    else
        exit
    fi
fi
TMHOME=`pwd`/chain tendermint init