table "log" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "user_id" {
    null = false
    type = uuid
  }
  column "action" {
    null = false
    type = character_varying(64)
  }
  column "status" {
    null = false
    type = text
  }
  column "time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "log_users_id_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "roles" {
  schema = schema.public
  column "id" {
    null = false
    type = character_varying(16)
  }
  primary_key {
    columns = [column.id]
  }
}
table "tokens" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = uuid
  }
  column "token" {
    null = false
    type = uuid
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "last_active_at" {
    null = false
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "tokens_users_id_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "user_roles" {
  schema = schema.public
  column "user_id" {
    null = false
    type = uuid
  }
  column "role" {
    null = false
    type = character(16)
  }
  primary_key {
    columns = [column.user_id]
  }
  foreign_key "user_roles_roles_id_fk" {
    columns     = [column.role]
    ref_columns = [table.roles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "user_roles_users_id_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "username" {
    null = false
    type = character_varying(64)
  }
  column "email" {
    null = false
    type = character_varying(320)
  }
  column "password_hash" {
    null = false
    type = character(64)
  }
  column "password_salt" {
    null = false
    type = character(16)
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "users_unique_email" {
    unique  = true
    columns = [column.email]
  }
  index "users_unique_username" {
    unique  = true
    columns = [column.username]
  }
}
schema "public" {
}
