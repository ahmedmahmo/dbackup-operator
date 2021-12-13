#!/usr/bin/env bash


#!/bin/bash
set -e
  
psql -v ON_ERROR_STOP=1 --username postgres --dbname postgres <<-EOSQL
    create schema test_schema;
  
    create table test_schema.employee(
        id  SERIAL PRIMARY KEY,
        firstname   TEXT    NOT NULL,
        lastname    TEXT    NOT NULL,
        email       TEXT    not null,
        age         INT     NOT NULL,
        salary         real,
        unique(email)
    );
  
    insert into test_schema.employee (firstname,lastname,email,age,salary)
    values ('John','Doe 1','john1@doe.com',18,1234.23);
 
EOSQL