create table bonus_types (
    id bigserial not null primary key,
    name_id text,
    name_en text,
    name_fr text,
    name_es text,
    name_de text,
    name_it text,
    name_pt text,
    created_at timestamp
    with
        time zone default now (),
        updated_at timestamp
    with
        time zone default now (),
        deleted_at timestamp
    with
        time zone
);

create unique index idx_bonus_types_name_id on bonus_types (name_id);

create table bonus (
    id bigserial not null primary key,
    bonus_type_id bigint not null references bonus_types (id),
    description_en text,
    description_fr text,
    description_es text,
    description_de text,
    description_it text,
    description_pt text,
    created_at timestamp
    with
        time zone default now (),
        updated_at timestamp
    with
        time zone default now (),
        deleted_at timestamp
    with
        time zone
);

create table tribute (
    id bigserial not null primary key,
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
    item_ankama_id bigint not null,
    item_subtype text not null,
    /* item category */
    item_doduapi_uri text not null,
    /* URI to the dofusdude api */
    quantity bigint not null,
    created_at timestamp
    with
        time zone default now (),
        updated_at timestamp
    with
        time zone default now (),
        deleted_at timestamp
    with
        time zone
);

create table almanax (
    id bigserial not null primary key,
    bonus_id bigint not null references bonus (id),
    tribute_id bigint not null references tribute (id),
    date text not null,
    reward_kamas bigint,
    created_at timestamp
    with
        time zone default now (),
        updated_at timestamp
    with
        time zone default now (),
        deleted_at timestamp
    with
        time zone
);

create unique index idx_almanax_date on almanax (date);
