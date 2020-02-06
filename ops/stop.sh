#!/usr/bin/env bash
docker service rm authority
docker service rm provider-{0,1,2}
docker service rm node-{0,1,2,3,4,5}
