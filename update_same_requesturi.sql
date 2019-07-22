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
