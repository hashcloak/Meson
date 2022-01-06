# /bin/sh

echo "Clean up data generated when execute scripts..."
rm -rf conf/node1/kdata/*
rm -rf conf/node1/data/*
rm -rf conf/node1/katzenmint
echo "{
  \"height\": \"0\",
  \"round\": 0,
  \"step\": 0
}" > conf/node1/data/priv_validator_state.json

rm -rf conf/node2/kdata/*
rm -rf conf/node2/data/*
rm -rf conf/node2/katzenmint
echo "{
  \"height\": \"0\",
  \"round\": 0,
  \"step\": 0
}" > conf/node2/data/priv_validator_state.json

rm -rf conf/node3/kdata/*
rm -rf conf/node3/data/*
rm -rf conf/node3/katzenmint
echo "{
  \"height\": \"0\",
  \"round\": 0,
  \"step\": 0
}" > conf/node3/data/priv_validator_state.json

rm -rf conf/mix1/data/*
rm -rf conf/mix2/data/*
rm -rf conf/mix3/data/*
rm -rf conf/provider1/data/*
rm -rf conf/provider2/data/*

# Update genesis block time
NOW=$(TZ=UTC date +"%Y-%m-%dT%H:%M:%S.000000Z")
TITLE="\"genesis_time\""
perl -i -pe"s/$TITLE.*/$TITLE: \"$NOW\",/g" conf/node1/config/genesis.json
perl -i -pe"s/$TITLE.*/$TITLE: \"$NOW\",/g" conf/node2/config/genesis.json
perl -i -pe"s/$TITLE.*/$TITLE: \"$NOW\",/g" conf/node3/config/genesis.json

echo "Cleaned up!"
