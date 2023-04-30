create table roles
(
    id   integer not null
        constraint roles_pk
            primary key autoincrement,
    role text    not null
);
insert into roles (id, role) values (1, 'admin'), (2, 'user');

create table users
(
    id         text   not null
        constraint users_pk
            primary key,
    first_name text      not null,
    username   text      not null,
    password_hash text   not null,
    password_salt text   not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp not null
);

create table user_roles
(
    user_id text not null
        constraint user_roles_users_id_fk
            references users,
    role_id integer not null
        constraint user_roles_roles_id_fk
            references roles,
    constraint user_roles_pk
        primary key (role_id, user_id)
);

