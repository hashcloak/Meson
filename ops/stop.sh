#!/usr/bin/env bash
docker service rm authority || true
docker service rm provider-{0,1,2} || true
docker service rm node-{0,1,2,3,4,5} || true
