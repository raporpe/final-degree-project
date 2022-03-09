-- PROBE REQUEST

-- Ver los verdaderos dispositivos encontrados
select count(distinct station_mac)from probe_request_frames where station_mac_vendor is not null;

-- Ver los verdaderos dispositivos encontrados con mac verdadera en 5ghz
select count(distinct station_mac)from probe_request_frames where station_mac_vendor is not null and frequency > 3000;

-- Ver los probe request capturados con mac verdadera
select count(*) from probe_request_frames where station_mac_vendor is not null;


-- Ver los dataframes capturados con mac verdadera
select count(distinct station_mac) from data_frames where station_mac_vendor is not null;


-- CONTROL FRAMES

-- Ver cada tipo de control en los control frame
select count(*) from control_frames where subtype = '4';

-- Porcentaje de cada tipo de control frame
select (count(*)::decimal/7800000)*100 from control_frames where subtype = '13';



-- DATA FRAMES

-- Ver porcentaje por cada tipo de data frame
select (count(*)::decimal/52000)*100 from data_frames where subtype = '0';

-- Variabilidad por cada subtipo de data frame
select count(distinct station_mac) from data_frames where subtype = '12';

-- Porcentaje de macs verdaderas por cada subtipo de data frame
select (b.a::decimal/b.total)*100 from (
    select
       (select count(distinct station_mac) from data_frames where subtype = '12') as total,
       (select count(distinct station_mac) from data_frames where subtype = '12' and station_mac_vendor is not null) as a
    ) as b;


-- Paquetes de datos con mac verdadera
select count(distinct station_mac) from data_frames where station_mac_vendor is not null;


-- MANAGEMENT FRAMES
select (count(*)::decimal/4900000)*100 from management_frames where subtype = '13';


-- OTROS

-- Comprobar que todos los bssid en data_frames tienen un equivalente en beacon_frames

(select d.station_mac, bf.ssid from data_frames as d
inner join
    beacon_frames bf on d.station_mac = bf.bssid)

select distinct bssid from beacon_frames;

-- Tamaño de la base de datos
select pg_size_pretty(pg_database_size('tfg'));


-- Estudio de los probe request

select (station_mac, count(*)) from probe_request_frames where intent is null and station_mac_vendor is null group by (station_mac);
select (station_mac, count(*)) from probe_request_frames where intent is null and station_mac_vendor is not null group by (station_mac);
select (station_mac, count(*)) from probe_request_frames where intent is not null and station_mac_vendor is not null group by (station_mac);
select (station_mac, count(*)) from probe_request_frames where intent is not null and station_mac_vendor is null group by (station_mac);


