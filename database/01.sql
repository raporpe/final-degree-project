create table data(
  id uuid not null default uuid_generate_v4() primary key,
  device_id varchar(64),
  station_bssid macaddr not null,
  ap_ssid macaddr,
  intent varchar(32),
  time int not null,
  power int
);