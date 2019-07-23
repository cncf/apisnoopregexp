-- this matches distinct requesturis and verbs and then populate this on the corresponing auditids
-- so this can update a lot of audits that share the same requesturi and verb - this is the most useful option IMHO
with ndata as (
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
  limit
    NNN
), data as (
  select distinct * from ndata
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
      d.requesturi = ev.requesturi
      and d.verb = ev.verb
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
      d.requesturi = ev.requesturi
      and d.verb = ev.verb
  ) >= 1 
;
