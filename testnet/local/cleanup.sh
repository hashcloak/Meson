# /bin/sh

echo "Clean up data generated when execute scripts..."

NOW=$(TZ=UTC date +"%Y-%m-%dT%H:%M:%S.000000Z")
TITLE="\"genesis_time\""

clean_mix_dir () {
    d=$1
    if [ -d $d ]; then
        echo "Clean mix data: $d"
        if [ -d $d/data ]; then
            rm -rf $d/data/*
        fi;
    fi;
}

clean_provider_dir () {
    d=$1
    if [ -d $d ]; then
        echo "Clean provider data: $d"
        if [ -d $d/data ]; then
            rm -rf $d/data/*
        fi;
    fi;
}

clean_katzenmint_dir () {
    d=$1
    if [ -d $d ]; then
        echo "Clean katzenmint data: $d"
        if [ ! -d $d/data ]; then
            mkdir $d/data
        fi;
        if [ -d $d/data ]; then
            rm -rf $d/data/*
        fi;
        if [ -d $d/kdata ]; then
            rm -rf $d/kdata/*
        fi;
        if [ -d $d/katzenmint ]; then
            rm -rf $d/katzenmint
        fi;
        echo "{
          \"height\": \"0\",
          \"round\": 0,
          \"step\": 0
        }" > $d/data/priv_validator_state.json
        # Update genesis block time
        perl -i -pe"s/$TITLE.*/$TITLE: \"$NOW\",/g" $d/config/genesis.json
    fi;
}

for d in conf/* ; do
    [ -L "${d%/}" ] && continue
    [[ $d =~ ^conf/mix[1-3] ]] && clean_mix_dir $d
    [[ $d =~ ^conf/provider[1-3] ]] && clean_provider_dir $d
    [[ $d =~ ^conf/node[1-4] ]] && clean_katzenmint_dir $d
done

echo "Cleaned up!"
