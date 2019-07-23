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
- Run: `make`, `sudo -u postgres ./rmatch`. That will update `op_id` column on `audit_events` table.


# Restore into the original database

- Generate SQL file to be run on the original database: `make`, `sudo -u postgres ./gensql > update.sql`.
- Run script on the original database: `` psql -h apisnoop-db-host -U apisnoop-db-user hh < update.sql ``.

You can also generate TSV dumps from your local databae and restore them on the remote, but the above solution is faster and better

- Dump `api_operations` table data into a local TSV file: `sudo -u postgres psql hh -tAc "\copy (select * from api_operations) to '/tmp/new_api_operations.tsv'"`.
- Dump `audit_events` table data into a local TSV file: `sudo -u postgres psql hh -tAc "\copy (select * from audit_events) to '/tmp/new_audit_events.tsv'"`.
