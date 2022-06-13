create table detected_macs (
    id uuid not null primary key,
    device_id text,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    detected_macs text,
    seconds_per_window bigint
);
create table personal_macs (
    mac text not null primary key,
    metadata text
);
create table rooms (
    device_id varchar not null constraint rooms_pk primary key,
    room_id varchar
);
create table room_historic (
    id varchar,
    data varchar,
    date timestamp with time zone
);
create table detected_macs_fix (
    id uuid,
    device_id text,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    detected_macs text,
    seconds_per_window bigint
);