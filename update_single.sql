-- first try to update record by finding another record with the same 'requesturi' and 'verb' fields
update
  audit_events e
set
  opid = (
    select
      i.opid
    from
      audit_events i
    where
      i.requesturi = e.requesturi
      and i.verb = e.verb
      and i.opid is not null
    limit
      1
  )
where
  e.opid is null
  and e.auditid = '8f0ef08b-01e5-47e0-8692-b14bf7754235'
  and (
    select
      count(*)
    from
      audit_events i
    where
      i.requesturi = e.requesturi
      and i.verb = e.verb
      and i.opid is not null
  ) >= 1
;
-- if previous operation success then 'opid' is non-null and next operations do nothing
-- now try to update by matching requesturi with regexp and method-verb mapping
with data as(
  select
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
    and ev.auditid = '8f0ef08b-01e5-47e0-8692-b14bf7754235'
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
    limit 1
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
