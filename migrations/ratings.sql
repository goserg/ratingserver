create table players
(
	id text not null
		constraint players_pk
			primary key,
	name text not null
		unique,
	created_at Timestamp default CURRENT_TIMESTAMP not null
);

create table matches
(
	id integer not null
		constraint matches_pk
			primary key autoincrement,
	player_a TEXT not null
		constraint matches_players_id_fk
			references players
				on delete cascade,
	player_b TEXT not null
		constraint matches_players_id_fk_2
			references players
				on delete cascade,
	winner TEXT
		constraint matches_players_id_fk_3
			references players
				on delete cascade,
	created_at timestamp default CURRENT_TIMESTAMP not null
);

