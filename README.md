# apisnoopregexp

# Clone API snoop database

- Create local postgres database: `sudo -u postgres psql`: `create database hh;`.
- Initialize database structure: `sudo -u postgres psql hh < structure.sql`.
- Dump `api_operations` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from api_operations) to 'api_operations.tsv'`.
- Dump `audit_events` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from audit_events) to 'audit_events.tsv'`.
- Restore `api_operations` table data locally: `sudo -u postgres psql hh -tAc "\copy api_operations from 'api_operations.tsv'"`.
- Restore `audit_events` table data locally: `sudo -u postgres psql hh -tAc "\copy audit_events from 'audit_events.tsv'"`.


# Update OpID via Regexp

- Run: `sudo -u postgres psql hh -c "`cat update_all.sql`"` to update 1000 entries on `audit_events` table (localize each request URI's matching `api_operations` `OpID`.
- Run: `sudo -u postgres psql hh -c "`cat update_same_requesturi.sql`"` to update all entries with `requesturi` that already have `OpID` identified in `audit_events` table.
- Run: `sudo -u postgres psql hh -c "`cat update_single.sql`"` to update single entry in `audit_events` table, replace hardcoded `8f0ef08b-01e5-47e0-8692-b14bf7754235` in `update_single.sql` file.
