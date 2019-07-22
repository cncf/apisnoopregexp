-- update audit_events set opid = null where opid is not null;
with data as(
  select
    op.id,
    ev.requesturi,
    ev.verb
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
  limit 1000
)
update
  audit_events ev
set
  opid = (
    select distinct
      d.id
    from
      data d
    where
      d.requesturi = ev.requesturi
      and d.verb = ev.verb
  )
where
  ev.opid is null
  and (
    select
      count(*)
    from
      data d
    where
      d.requesturi = ev.requesturi
      and d.verb = ev.verb
  ) >= 1 
;
