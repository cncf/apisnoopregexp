-- To rollback all updates
-- update audit_events set op_id = null where op_id is not null;
-- updates using audit_id (because there are much more audit_ids than request_uris and verbs, you shoudl run update_same_requesturi.sql after this)
with ndata as (
  select
    op.id,
    ev.audit_id
  from
    api_operations op,
    audit_events ev
  where
    ev.op_id is null
    and (
      (op.method = 'get' and ev.verb in ('get', 'list', 'proxy'))
      or (op.method = 'patch' and ev.verb = 'patch')
      or (op.method = 'put' and ev.verb = 'update')
      or (op.method = 'post' and ev.verb = 'create')
      or (op.method = 'delete' and ev.verb in ('delete', 'deletecollection'))
      or (op.method = 'watch' and ev.verb in ('watch', 'watchlist'))
    )
    and ev.request_uri ~ op.regexp
  limit
    NNN
), data as (
  select distinct * from ndata
)
update
  audit_events ev
set
  op_id = (
    select
      d.id
    from
      data d
    where
      d.audit_id = ev.audit_id
    limit
      1
  )
where
  ev.op_id is null
  and (
    select
      count(*)
    from
      data d
    where
      d.audit_id = ev.audit_id
  ) >= 1 
;
-- select * from data;
/*select 
 d.audit_id,
 d.id
from
  audit_events ae,
  data d
where
  d.op_id is null
  and ae.audit_id = d.audit_id
order by
  d.audit_id,
  d.id
;*/
