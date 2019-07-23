-- first try to update record by finding another record with the same 'request_uri' and 'verb' fields
update
  audit_events e
set
  op_id = (
    select
      i.op_id
    from
      audit_events i
    where
      i.request_uri = e.request_uri
      and i.verb = e.verb
      and i.op_id is not null
    limit
      1
  )
where
  e.op_id is null
  and e.audit_id = '8f0ef08b-01e5-47e0-8692-b14bf7754235'
  and (
    select
      count(*)
    from
      audit_events i
    where
      i.request_uri = e.request_uri
      and i.verb = e.verb
      and i.op_id is not null
  ) >= 1
;
-- if previous operation success then 'op_id' is non-null and next operations do nothing
-- now try to update by matching request_uri with regexp and method-verb mapping
with data as(
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
    and ev.audit_id = '8f0ef08b-01e5-47e0-8692-b14bf7754235'
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
    limit 1
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
