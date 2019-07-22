update
  audit_events e
set
  opid = (
    select distinct
      i.opid
    from
      audit_events i
    where
      i.requesturi = e.requesturi
      and i.verb = e.verb
      and i.opid is not null
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
