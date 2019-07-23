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
