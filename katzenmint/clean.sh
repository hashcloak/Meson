# /bin/sh
rm -rf chain/data/*
echo "{
  \"height\": \"0\",
  \"round\": 0,
  \"step\": 0
}" > chain/data/priv_validator_state.json

rm -rf data/*