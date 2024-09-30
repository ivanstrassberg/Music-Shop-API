create extension if not exists pgcrypto; 

create table if not exists "songs"  (
    uuid UUID default gen_random_uuid() primary key,
    song_name varchar(255) not null,
    group_name varchar(255) not null,
    release_date date,
    song_text text,
    song_link varchar(255),
    created_at timestamp default current_timestamp not null,
    updated_at timestamp default null
);
