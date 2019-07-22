# apisnoopregexp

# Clone API snoop database

- Create local postgres database: `sudo -u postgres psql`: `create database hh;`.
- Initialize database structure: `sudo -u postgres psql hh < structure.sql`.
- Dump `api_operations` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from api_operations) to '/tmp/api_operations.tsv'"`.
- Dump `audit_events` table data into a local TSV file: `psql -h apisnoop-db-host -U apisnoop-db-user hh -tAc "\copy (select * from audit_events) to '/tmp/audit_events.tsv'"`.
- Restore `api_operations` table data locally: `sudo -u postgres psql hh -tAc "\copy api_operations from '/tmp/api_operations.tsv'"`.
- Restore `audit_events` table data locally: `sudo -u postgres psql hh -tAc "\copy audit_events from '/tmp/audit_events.tsv'"`.


# Update OpID via Regexp

Note that all mass update script only update records with `null` on `opid`, so they can be called iteratively.

- Run: `sudo -u postgres psql hh -c "`cat update_by_requesturi.sql`"` to update 1000 `requesturi`/`verb` entries on `audit_events` table (localize each request URI's matching `api_operations` `OpID`).
- Run: `sudo -u postgres psql hh -c "`cat update_single.sql`"` to update single entry in `audit_events` table, replace hardcoded `8f0ef08b-01e5-47e0-8692-b14bf7754235` in `update_single.sql` file.
- You can runn  `./loop.sh` script to add regexp matching in 100 elements packs one by one until no more updates is made (they're packs of 100 different `requesturi`/`verb` packs). There are about 720 such packs.


# Restore into the original database

- Dump `api_operations` table data into a local TSV file: `sudo -u postgres psql hh -tAc "\copy (select * from api_operations) to '/tmp/new_api_operations.tsv'"`.
- Dump `audit_events` table data into a local TSV file: `sudo -u postgres psql hh -tAc "\copy (select * from audit_events) to '/tmp/new_audit_events.tsv'"`.


# Different approach

Some suboptimal calls, just for reference:

- Run: `sudo -u postgres psql hh -c "`cat update_by_auditid.sql`"` to update 1000 `auditid` entries on `audit_events` table (localize each request URI's matching `api_operations` `OpID`).
- Run: `sudo -u postgres psql hh -c "`cat update_same_requesturi.sql`"` to update all entries with `requesturi`/`verb` that already have `OpID` identified in `audit_events` table (not using regexp matching).
