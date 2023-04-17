create table main.user_players
(
    user_id   integer not null
        constraint user_players_pk
            primary key
        constraint user_players_users_id_fk
            references users,
    player_id text    not null
);
