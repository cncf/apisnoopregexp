create table api_operations(
  id text not null,
  method text not null,
  path text not null,
  regexp text not null,
  "group" text not null,
  version text not null,
  kind text not null,
  category text not null,
  description text not null
);

create table audit_events(
  auditID            uuid,
  testrunID          text,
  opID               text,
  level              text not null,
  verb               text not null,
  requestURI         text not null,
  userAgent          text,
  testName           text,
  requestkind        text not null,
  requestapiversion  text not null,
  requestmeta        jsonb not null,
  requestspec        jsonb not null,
  requeststatus      jsonb not null,
  responsekind       text not null,
  responseapiversion text not null,
  responsemeta       jsonb not null,
  responsespec       jsonb not null,
  responsestatus     jsonb not null,
  timeStamp          timestamp with time zone
);

-- Indexes
create index api_operations_id on api_operations(id);
create index api_operations_method on api_operations(method);
create index api_operations_regexp on api_operations(regexp);

create index audit_events_opid on audit_events(opid);
create index audit_events_verb on audit_events(verb);
create index audit_events_requesturi on audit_events(requesturi);
