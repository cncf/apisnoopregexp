-- To rollback all updates
-- update audit_events set opid = null where opid is not null;
-- updates using auditid (becauser there are much more auditids than requesturis and verbs, you shoudl run update_same_requesturi.sql after this)
with data as(
  select distinct
    op.id,
    ev.auditid
  from
    api_operations op,
    audit_events ev
  where
    ev.opid is null
    and (
      (op.method = 'get' and ev.verb in ('get', 'list', 'proxy'))
      or (op.method = 'patch' and ev.verb = 'patch')
      or (op.method = 'put' and ev.verb = 'update')
      or (op.method = 'post' and ev.verb = 'create')
      or (op.method = 'delete' and ev.verb in ('delete', 'deletecollection'))
      or (op.method = 'watch' and ev.verb in ('watch', 'watchlist'))
    )
    and ev.requesturi ~ op.regexp
  limit
    3
)
update
  audit_events ev
set
  opid = (
    select
      d.id
    from
      data d
    where
      d.auditid = ev.auditid
    limit
      1
  )
where
  ev.opid is null
  and (
    select
      count(*)
    from
      data d
    where
      d.auditid = ev.auditid
  ) >= 1 
;
-- select * from data;
/*select 
 d.auditid,
 d.id
from
  audit_events ae,
  data d
where
  d.opid is null
  and ae.auditid = d.auditid
order by
  d.auditid,
  d.id
;*/
