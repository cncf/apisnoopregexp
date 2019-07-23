#!/bin/bash
# NOINF=1 (skip info)
ts=`date +'%s%N'`
fn="/tmp/$ts.sql"
function finish {
  rm -f "$fn" 2>/dev/null
}
trap finish EXIT
if [ -z "$1" ]
then
  n=500
else
  n=$1
fi
echo "$fn"
# cp update_by_auditid.sql "$fn"
cp update_by_requesturi.sql "$fn"
vim --not-a-term -c "%s/NNN/${n}/g" -c 'wq!' "$fn"
while true
do
  if [ -z "$NOINF" ]
  then
    res=`sudo -u postgres psql hh -tAc 'select count(distinct (requesturi || verb)) filter (where opid is not null) as found, count(distinct (requesturi || verb)) filter (where opid is null) as not_found from audit_events'`
    echo "Found|NotFound: $res"
  fi
  echo "Processing next $n..."
  sudo -u postgres psql hh -tAc "`cat $fn`" > out
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
