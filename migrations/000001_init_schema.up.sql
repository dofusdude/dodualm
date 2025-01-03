create table bonus_types (
    id integer primary key autoincrement,
    name_id text,
    name_en text,
    name_fr text,
    name_es text,
    name_de text,
    name_it text,
    name_pt text,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime
);

create unique index idx_bonus_types_name_id on bonus_types (name_id);

create table bonus (
    id integer primary key autoincrement,
    bonus_type_id integer not null,
    description_fr text,
    description_es text,
    description_de text,
    description_it text,
    description_pt text,
    description_en text,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime,
    foreign key (bonus_type_id) references bonus_types (id)
);

create table tribute (
    id integer primary key autoincrement,
    item_name_en text,
    item_name_fr text,
    item_name_es text,
    item_name_de text,
    item_name_it text,
    item_name_pt text,
    item_icon text not null,
    item_sd text,
    item_hq text,
    item_hd text,
    item_ankama_id integer not null,
    item_subtype text not null,
    /* item category */
    item_doduapi_uri text not null,
    /* URI to the dofusdude api */
    quantity integer not null,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime
);

create table almanax (
    id integer not null primary key autoincrement,
    bonus_id integer not null,
    tribute_id integer not null,
    date text not null,
    reward_kamas integer,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime,
    foreign key (bonus_id) references bonus (id),
    foreign key (tribute_id) references tribute (id)
);

create unique index idx_almanax_date on almanax (date);
