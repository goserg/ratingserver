create table roles
(
    id   integer not null
        constraint roles_pk
            primary key autoincrement,
    role text    not null
);

create table users
(
    id         integer   not null
        constraint users_pk
            primary key autoincrement,
    first_name text      not null,
    username   text      not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

create table log
(
    id      integer not null
        constraint log_pk
            primary key autoincrement,
    user_id integer not null
        constraint log_users_id_fk
            references users,
    message integer
);

create table user_roles
(
    user_id integer not null
        constraint user_roles_users_id_fk
            references users,
    role_id integer not null
        constraint user_roles_roles_id_fk
            references roles,
    constraint user_roles_pk
        primary key (role_id, user_id)
);
