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
  audit_id           uuid, -- changed
  testrun_id         text, --changed
  op_id              text, -- changed
  stage              text not null, -- new
  level              text not null,
  verb               text not null,
  request_uri        text not null, -- changed
  user_agent         text, -- changed
  test_name          text, -- changed
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
  request_ts         timestamp with time zone, --new
  stage_ts           timestamp with time zone --new
);

-- Indexes
create index api_operations_id on api_operations(id);
create index api_operations_method on api_operations(method);
create index api_operations_regexp on api_operations(regexp);

create index audit_events_op_id on audit_events(op_id);
create index audit_events_verb on audit_events(verb);
create index audit_events_request_uri on audit_events(request_uri);
