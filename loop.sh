#!/bin/bash
while true
do
  res=`sudo -u postgres psql hh -tAc 'select count(distinct (requesturi || verb)) filter (where opid is not null) as found, count(distinct (requesturi || verb)) filter (where opid is null) as not_found from audit_events'`
  echo "Found|NotFound: $res"
  echo "Processing next 500..."
  res=''
  # sudo -u postgres psql hh -tA < update_by_requesturi.sql > out
  sudo -u postgres psql hh -tAc "`cat update_by_requesturi.sql`" > out
  res=`cat out`
  if [ -z "$res" ]
  then
    echo "Error: no results, exiting"
    exit 1
  fi
  echo "$res"
  if [ "$res" = "UPDATE 0" ]
  then
    echo "Finished."
    exit 0
  fi
  sleep 1
done
