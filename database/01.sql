create table probe_request_frames(
    id uuid not null default uuid_generate_v4() primary key,
    device_id varchar(32) not null,
    station_mac macaddr not null,
    intent varchar(32),
    time int not null,
    power smallint not null,
    station_mac_vendor varchar(128)
);


create table probe_response_frames (
    id uuid not null default uuid_generate_v4() primary key,
    device_id varchar(32) not null,
    bssid macaddr not null,
    ssid varchar(32),
    station_mac macaddr not null,
    station_mac_vendor varchar(128),
    time int not null,
    power smallint not null
);

create table beacon_frames (
    bssid macaddr primary key,
    ssid varchar(32),
    device_id varchar(32) not null
);

create table data_frames (
    id uuid not null default uuid_generate_v4() primary key,
    bssid macaddr not null,
    station_mac macaddr not null,
    time int not null,
    power smallint not null,
    station_mac_vendor varchar(128)
);

create table control_frames (
    id uuid not null default uuid_generate_v4() primary key,
    bssid macaddr not null,
    station_mac macaddr not null,
    subtype varchar(32) not null,
    time int not null,
    power smallint not null,
    station_mac_vendor varchar(128)
);

create table management_frames (
    id uuid not null default uuid_generate_v4() primary key,
    addr1 macaddr not null,
    addr2 macaddr,
    addr3 macaddr,
    addr4 macaddr,
    time int not null,
    subtype varchar(32) not null,
    power smallint not null
);


create table ocupation (
    id uuid default uuid_generate_v4() primary key,
    device_id varchar,
    count int
);