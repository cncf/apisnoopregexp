#!/bin/bash
while true
do
  res=`sudo -u postgres psql hh -tAc 'select count(distinct (requesturi || verb)) filter (where opid is not null) as found, count(distinct (requesturi || verb)) filter (where opid is null) as not_found from audit_events'`
  echo "Found|NotFound: $res"
  echo "Processing next 100..."
  # res=`sudo -u postgres psql hh -tAc "update audit_events set opid = 'xxx' where requesturi = 'yyy'"`
  res=`sudo -u postgres psql hh -tAc "`cat update_by_requesturi.sql`"`
  echo "Updated: $res"
  if [ "$res" = "UPDATE 0" ]
  then
    echo "Finished."
    exit 0
  fi
  sleep 1
done
