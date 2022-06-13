create table "old-tfg".beacon_frames (
    bssid macaddr not null constraint access_point_pkey primary key,
    ssid varchar(32),
    device_id varchar(32),
    frequency smallint
);
create table "old-tfg".control_frames (
    id uuid default uuid_generate_v4() not null constraint action_frames_pkey primary key,
    addr1 macaddr not null,
    time timestamp not null,
    power smallint not null,
    subtype varchar(32),
    addr2 macaddr,
    addr3 macaddr,
    addr4 macaddr,
    frequency smallint
);
create table "old-tfg".data_frames (
    id uuid default uuid_generate_v4() not null constraint dataframes_pkey primary key,
    bssid macaddr not null,
    station_mac macaddr not null,
    time timestamp not null,
    power smallint not null,
    station_mac_vendor varchar(128),
    subtype varchar(32),
    frequency smallint
);
create table "old-tfg".management_frames (
    id uuid default uuid_generate_v4() not null primary key,
    addr1 macaddr,
    addr2 macaddr,
    addr3 macaddr,
    addr4 macaddr,
    time timestamp not null,
    subtype varchar(32) not null,
    power smallint not null,
    frequency smallint
);
create table "old-tfg".probe_request_frames (
    id uuid default uuid_generate_v4() not null constraint data_pkey primary key,
    device_id varchar(32),
    station_mac macaddr not null,
    intent varchar(32),
    time timestamp not null,
    power smallint,
    station_mac_vendor varchar(128),
    frequency smallint
);
create table "old-tfg".probe_response_frames (
    id uuid default uuid_generate_v4() not null primary key,
    device_id varchar(32) not null,
    bssid macaddr not null,
    ssid varchar(32),
    station_mac macaddr not null,
    station_mac_vendor varchar(128),
    time timestamp not null,
    power smallint not null,
    frequency smallint
);
