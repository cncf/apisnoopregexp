# apisnoopregexp

# Clone API snoop database

- Create local postgres database: `sudo -u postgres psql`: `create database hh;`.
- Initialize database structure: `sudo -u postgres psql hh < structure.sql`.
- Dump `api_operations` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from api_operations) to 'api_operations.tsv'`.
- Dump `audit_events` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from audit_events) to 'audit_events.tsv'`.
- Restore `api_operations` table data locally: `sudo -u postgres psql hh -tAc "\copy api_operations from 'api_operations.tsv'"`.
- Restore `audit_events` table data locally: `sudo -u postgres psql hh -tAc "\copy audit_events from 'audit_events.tsv'"`.
