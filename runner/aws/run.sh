#!/usr/bin/env bash


DIR="/tmp"
TIME=$(date +%s)
USERNAME="postgres"
PASSWORD="dummydummy"


pg_dump --no-owner --dbname=postgresql://postgres:dummydummy@postgres.postgres.svc.cluster.local:5432/postgres > file.sql  

sleep 100